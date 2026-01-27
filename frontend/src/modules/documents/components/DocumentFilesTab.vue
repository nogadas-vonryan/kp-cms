<template>
  <div class="space-y-4">
    <UiCard>  
      <div 
        class="mb-4 border-2 border-dashed rounded-lg p-8 transition-colors duration-200 text-center"
        :class="[
          isDragging ? 'border-blue-400 bg-blue-50' : 'border-gray-200 bg-gray-50',
          isUploading ? 'opacity-50 pointer-events-none' : 'cursor-pointer',
        ]"
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="handleDrop"
        @click="fileInput?.click()"
      >
        <input type="file" ref="fileInput" class="hidden" @change="handleFileSelect" />
        <div class="space-y-1">
          <p class="text-sm font-medium text-gray-700 uppercase tracking-wide">
            {{ isUploading ? 'Uploading...' : 'Click or drag to upload' }}
          </p>
          <p class="text-[11px] text-gray-400 uppercase tracking-widest">Max 32MB per file</p>
        </div>
      </div>

      <div v-if="document.files?.length > 0" class="mb-6 space-y-2">
        <label class="block text-sm font-medium text-gray-700">Filter Files</label>
        <div class="space-y-3">
          <input 
            v-model="searchQuery"
            type="text" 
            placeholder="Search by name or description..."
            class="w-full bg-white border border-gray-300 text-sm p-2 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
          <div class="flex gap-2 overflow-x-auto whitespace-nowrap py-1">
            <button
              v-for="tag in availableTags"
              :key="tag"
              type="button"
              @click="toggleTag(tag)"
              class="px-3 py-1.5 text-xs font-semibold rounded-full border transition-colors shrink-0"
              :class="selectedTags.includes(tag)
                ? 'bg-blue-600 text-white border-blue-600 shadow-sm'
                : 'bg-white text-gray-700 border-gray-300 hover:border-blue-400 hover:text-blue-700'"
            >
              {{ tag }}
            </button>
            <span v-if="availableTags.length === 0" class="text-xs text-gray-400 italic shrink-0">No tags available</span>
          </div>
          <div class="flex items-center justify-between text-[11px] text-gray-500 uppercase tracking-widest">
            <span>Search and/or pick tags; all selected tags must be present.</span>
            <button v-if="selectedTags.length" type="button" @click="clearTags" class="text-blue-600 hover:text-blue-800 font-semibold">Clear tags</button>
          </div>
        </div>
      </div>

      <UiSystemNotice
        v-model="statusMessage"
        :type="statusMessage?.type ?? 'success'"
        :title="statusMessage?.text"
      />

      <div class="space-y-0 divide-y divide-gray-200 border-t border-gray-200">
        <div v-if="filteredFiles.length === 0" class="py-12 text-center text-sm text-gray-500 italic">
          No files attached to this document.
        </div>

        <div 
          v-for="file in filteredFiles" 
          :key="file.file_name"
      :class="[
        'group py-4 transition-colors hover:bg-gray-50/50',
        highlightActive === file.file_name
          ? 'p-2 bg-yellow-200 shadow-lg'
          : ''
      ]"
			:ref="el => setRowRef(file.file_name, el as HTMLElement | null)"
        >
          <div class="flex justify-between items-start">
            <div class="flex-1 min-w-0 pr-4">
              <div class="flex items-center gap-2 mb-2">
                <span class="text-sm font-bold text-gray-900 truncate">{{ file.file_name }}</span>
                <span class="text-[11px] font-mono text-gray-400 uppercase">{{ formatSize(file.size) }}</span>
              </div>

              <div v-if="editingFileName !== file.file_name" class="grid grid-cols-1 sm:grid-cols-2 gap-x-4 gap-y-2">
                <div>
                  <label class="block text-xs font-medium text-gray-500 tracking-tighter">Description</label>
                  <div class="text-sm text-gray-700">{{ file.description || '—' }}</div>
                </div>
                <div v-if="file.note">
                  <label class="block text-xs font-medium text-gray-500 tracking-tighter">Note</label>
                  <div class="text-sm text-gray-700">{{ file.note }}</div>
                </div>
                <div v-if="file.tags && file.tags.length > 0">
                  <label class="block text-xs font-medium text-gray-500 tracking-tighter">Tags</label>
                  <div class="text-sm text-gray-700">{{ Array.isArray(file.tags) ? file.tags.join(', ') : file.tags }}</div>
                </div>
              </div>
            </div>

            <div class="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-all duration-200 whitespace-nowrap pt-1">
              <button 
                @click="downloadFile(file.file_name)" 
                class="px-3 py-1.5 text-[10px] font-bold uppercase tracking-widest text-blue-600 bg-blue-50 hover:bg-blue-600 hover:text-white rounded transition-colors"
              >
                Download
              </button>
              
              <template v-if="isAdmin">
                <button 
                  @click="startEdit(file)" 
                  class="px-3 py-1.5 text-[10px] font-bold uppercase tracking-widest text-gray-600 bg-gray-100 hover:bg-gray-800 hover:text-white rounded transition-colors"
                >
                  Edit
                </button>
                <button 
                  @click="deleteFile(file.file_name)" 
                  class="px-3 py-1.5 text-[10px] font-bold uppercase tracking-widest text-red-600 bg-red-50 hover:bg-red-600 hover:text-white rounded transition-colors"
                >
                  Delete
                </button>
              </template>
            </div>
          </div>

          <div v-if="editingFileName === file.file_name" class="mt-4 p-4 bg-gray-50 rounded border border-gray-200 space-y-4">
            <div>
              <label class="block text-xs font-bold text-gray-600 uppercase mb-1">Edit Description</label>
              <input v-model="editForm.description" class="w-full text-sm p-2 border bg-white border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-600 uppercase mb-1">Note</label>
              <textarea v-model="editForm.note" class="w-full text-sm p-2 border bg-white border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" rows="2"></textarea>
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-600 uppercase mb-2">Tags</label>
              <div class="space-y-2">
                <div v-for="(_, index) in editForm.tags" :key="index" class="flex gap-2">
                  <UiInput v-model="editForm.tags[index]" placeholder="Enter tag" class="flex-1" />
                  <button 
                    @click="editForm.tags.splice(index, 1)" 
                    class="px-3 py-2 text-xs font-bold text-red-600 bg-red-50 hover:bg-red-600 hover:text-white rounded transition-colors"
                  >
                    Remove
                  </button>
                </div>
                <button 
                  @click="editForm.tags.push('')" 
                  class="w-full px-3 py-2 text-xs font-bold text-blue-600 bg-blue-50 hover:bg-blue-100 rounded transition-colors"
                >
                  + Add Tag
                </button>
              </div>
            </div>
            <div class="flex justify-end gap-3 pt-2 border-t border-gray-200">
              <button @click="editingFileName = null" class="text-xs font-medium text-gray-500 hover:text-gray-700">Cancel</button>
              <button @click="saveMetadata(file.file_name)" class="text-xs font-bold bg-blue-600 text-white px-4 py-1.5 rounded hover:bg-blue-700">Save Changes</button>
            </div>
          </div>
        </div>
      </div>
    </UiCard>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue';
import { UiCard } from '@/core/ui';
import { DocumentService } from '@/modules/documents/services/documentService';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';
import UiInput from '@/core/ui/components/UiInput.vue';

const props = defineProps<{
  document: any;
  isAdmin: boolean;
	highlightFileName?: string;
}>();

const emit = defineEmits(['refresh']);

const fileInput = ref<HTMLInputElement | null>(null);
const isDragging = ref(false);
const isUploading = ref(false);
const statusMessage = ref<any>(null);
const highlightActive = ref('');
const pendingHighlight = ref('');
const highlightTimer = ref<ReturnType<typeof setTimeout> | null>(null);
const rowRefs = ref<Record<string, HTMLElement | null>>({});
const isHighlighting = ref(false);

const searchQuery = ref('');
const selectedTags = ref<(string | number)[]>([]);
const editingFileName = ref<string | null>(null);
const editForm = ref<{ description: string; note: string; tags: string[] }>({ description: '', note: '', tags: [] });

const availableTags = computed(() => {
  const tagSet = new Set<string>();
  (props.document.files ?? []).forEach((f: any) => {
    if (Array.isArray(f.tags)) {
      f.tags.forEach((tag: string | number) => {
        if (tag) tagSet.add(String(tag));
      });
    } else if (typeof f.tags === 'string' && f.tags.trim()) {
      tagSet.add(f.tags.trim());
    }
  });
  return Array.from(tagSet).sort((a, b) => a.localeCompare(b));
});

const filteredFiles = computed(() => {
  if (!props.document.files) return [];
  const q = searchQuery.value.toLowerCase();
  const selected = selectedTags.value.map((t) => String(t));

  return props.document.files.filter((f: any) => {
    const fileTags = Array.isArray(f.tags)
      ? f.tags.map((tag: string | number) => String(tag))
      : (typeof f.tags === 'string' && f.tags.trim())
        ? [f.tags.trim()]
        : [];

    const matchesSearch = [
      f.file_name,
      f.description,
      f.note,
      ...fileTags,
    ].some((value) => value && String(value).toLowerCase().includes(q));

    const matchesTags = selected.length === 0
      || selected.every(tag => fileTags.includes(tag));

    return matchesSearch && matchesTags;
  });
});

function setRowRef(fileName: string, el: HTMLElement | null) {
  if (el) {
    rowRefs.value[fileName] = el;
  } else {
    delete rowRefs.value[fileName];
  }
}

/**
 * Downloads file with proper authentication using blob.
 */
async function downloadFile(fileName: string) {
  try {
    const blob = await DocumentService.downloadFile(props.document.uuid, fileName);
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', fileName);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    window.URL.revokeObjectURL(url);
  } catch (error: any) {
    statusMessage.value = {
      type: 'error',
      text: getErrorMessage(error)
    };
  }
}

function startEdit(file: any) {
  editingFileName.value = file.file_name;
  const tags = Array.isArray(file.tags) ? [...file.tags] : [];
  editForm.value = { 
    description: file.description || '', 
    note: file.note || '',
    tags
  };
}

function toggleTag(tag: string) {
  const tagStr = String(tag);
  const idx = selectedTags.value.findIndex((t) => String(t) === tagStr);
  if (idx >= 0) {
    selectedTags.value.splice(idx, 1);
  } else {
    selectedTags.value.push(tagStr);
  }
}

function clearTags() {
  selectedTags.value = [];
}

async function saveMetadata(fileName: string) {
  try {
    // Filter out empty tags
    const metadata = {
      ...editForm.value,
      tags: editForm.value.tags.filter(t => t.trim())
    };
    
    await DocumentService.updateFileMetadata(props.document.uuid, fileName, metadata);
    editingFileName.value = null;
    emit('refresh');
    statusMessage.value = { type: 'success', text: 'metadata is successfully updated' };
  } catch (err: any) {
    statusMessage.value = { type: 'error', text: 'update failed' };
  }
}

async function deleteFile(fileName: string) {
  if (!confirm(`Confirm Deletion: ${fileName}`)) return;
  try {
    await DocumentService.deleteFile(props.document.uuid, fileName);
    emit('refresh');
  } catch (err: any) {
    statusMessage.value = { type: 'error', text: 'DELETE_FAILED' };
  }
}

const handleDrop = (e: DragEvent) => {
  isDragging.value = false;
  const file = e.dataTransfer?.files?.[0];
  if (file) uploadFile(file);
};

const handleFileSelect = (e: Event) => {
  const target = e.target as HTMLInputElement;
  const file = target.files?.[0];
  if (file) uploadFile(file);
};

const getErrorMessage = (err: any): string => {
  if (err.response?.data?.error) {
    const rawError = err.response.data.error;
    if (rawError.includes('file exists')) {
      return 'Conflict: File path already occupied in storage.';
    }
    return `API_${err.response.status}: ${rawError}`;
  }
  return err.message || 'Unknown System Exception';
};

async function uploadFile(file: File) {
  statusMessage.value = null;
  try {
    isUploading.value = true;
    
    const response = await DocumentService.uploadFile(props.document.uuid, file);
    
    // Handle Success
    statusMessage.value = {
      type: 'success',
      text: response.data.message || `${file.name} synchronized successfully.`
    };
    
    // Trigger parent refresh to update props.document.files list
    emit('refresh');
    
  } catch (error: any) {
    console.error('[System] Upload Failed:', error);
    statusMessage.value = {
      type: 'error',
      text: getErrorMessage(error)
    };
  } finally {
    isUploading.value = false;
    if (fileInput.value) fileInput.value.value = '';
  }
} 

function clearHighlightTimer() {
  if (highlightTimer.value) {
    clearTimeout(highlightTimer.value);
    highlightTimer.value = null;
  }
}

async function highlightAndScroll(fileName: string) {
  if (!fileName) return;
  if (isHighlighting.value) return;
  isHighlighting.value = true;

  const shouldClearFilters = searchQuery.value !== '' || selectedTags.value.length > 0;
  if (shouldClearFilters) {
    searchQuery.value = '';
    selectedTags.value = [];
    await nextTick();
  }

  await nextTick();
  const target = rowRefs.value[fileName];
  highlightActive.value = fileName;
  if (target) {
    target.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }
  clearHighlightTimer();
  highlightTimer.value = setTimeout(() => {
    highlightActive.value = '';
    highlightTimer.value = null;
  }, 4000);

  isHighlighting.value = false;
}

watch(
  () => props.highlightFileName,
  (fileName) => {
    if (!fileName) return;
    pendingHighlight.value = fileName;
    const exists = filteredFiles.value.some((f: any) => f.file_name === fileName);
    if (exists) highlightAndScroll(fileName);
  }
);

watch(filteredFiles, () => {
  if (isHighlighting.value) return;
  const target = pendingHighlight.value || props.highlightFileName || '';
  if (target && filteredFiles.value.some((f: any) => f.file_name === target)) {
    highlightAndScroll(target);
    pendingHighlight.value = '';
  }
});

function formatSize(bytes?: number) {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
</script>