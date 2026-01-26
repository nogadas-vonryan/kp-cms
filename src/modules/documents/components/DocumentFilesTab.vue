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

      <div v-if="document.files?.length > 0" class="mb-6">
        <label class="block text-sm font-medium text-gray-700 mb-1">Filter Files</label>
        <input 
          v-model="searchQuery"
          type="text" 
          placeholder="Search by name or description..."
          class="w-full bg-white border border-gray-300 text-sm p-2 rounded focus:ring-1 focus:ring-blue-500 outline-none"
        />
      </div>

      <UiSystemNotice v-model="statusMessage" :type="(statusMessage && statusMessage.type) ? statusMessage.type : 'success'" />

      <div class="space-y-0 divide-y divide-gray-200 border-t border-gray-200">
        <div v-if="filteredFiles.length === 0" class="py-12 text-center text-sm text-gray-500 italic">
          No files attached to this document.
        </div>

        <div 
          v-for="file in filteredFiles" 
          :key="file.file_name"
          class="group py-4 transition-colors hover:bg-gray-50/50"
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
              <input v-model="editForm.description" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-600 uppercase mb-1">Note</label>
              <textarea v-model="editForm.note" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" rows="2"></textarea>
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
import { ref, computed } from 'vue';
import { UiCard } from '@/core/ui';
import { DocumentService } from '@/modules/documents/services/documentService';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';

const props = defineProps<{
  document: any;
  isAdmin: boolean;
}>();

const emit = defineEmits(['refresh']);

const fileInput = ref<HTMLInputElement | null>(null);
const isDragging = ref(false);
const isUploading = ref(false);
const statusMessage = ref<any>(null);

const searchQuery = ref('');
const editingFileName = ref<string | null>(null);
const editForm = ref({ description: '', note: '' });

const filteredFiles = computed(() => {
  if (!props.document.files) return [];
  const q = searchQuery.value.toLowerCase();
  return props.document.files.filter((f: any) => 
    f.file_name.toLowerCase().includes(q) || 
    (f.description && f.description.toLowerCase().includes(q))
  );
});

/**
 * FIXED DOWNLOAD LOGIC
 * Creates a temporary link to force the browser to recognize the download stream.
 */
function downloadFile(fileName: string) {
  const url = DocumentService.getFileDownloadUrl(props.document.uuid, fileName);
  const link = document.createElement('a');
  link.href = url;
  // This attribute hints to the browser to download instead of navigate
  link.setAttribute('download', fileName);
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
}

function startEdit(file: any) {
  editingFileName.value = file.file_name;
  editForm.value = { description: file.description || '', note: file.note || '' };
}

async function saveMetadata(fileName: string) {
  try {
    await DocumentService.updateFileMetadata(props.document.uuid, fileName, editForm.value);
    editingFileName.value = null;
    emit('refresh');
    statusMessage.value = { type: 'success', text: 'METADATA_UPDATED' };
  } catch (err: any) {
    statusMessage.value = { type: 'error', text: 'UPDATE_FAILED' };
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
    
    // Perform upload via DocumentService
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

function formatSize(bytes?: number) {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
</script>