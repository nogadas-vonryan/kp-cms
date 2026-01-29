<template>
  <div class="space-y-4">
    <UiCard>  
      <div 
        class="mb-4 border-2 border-dashed rounded-lg p-4 sm:p-8 transition-colors duration-200 text-center"
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
          <p class="text-xs sm:text-sm font-medium text-gray-700 uppercase tracking-wide">
            {{ isUploading ? 'Uploading...' : 'Click or drag to upload' }}
          </p>
          <p class="text-[10px] sm:text-[11px] text-gray-400 uppercase tracking-widest">Max 32MB per file</p>
        </div>
      </div>

      <div v-if="document.files?.length > 0" class="mb-6 space-y-3">
        <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
          <label class="block text-sm font-medium text-gray-700">Filter Files</label>
          <div class="flex gap-0 border border-gray-300">
            <button
              @click="viewMode = 'list'"
              :title="viewMode === 'list' ? 'List view (active)' : 'List view'"
              :class="[
                'px-3 py-1.5 transition-colors flex items-center justify-center rounded',
                viewMode === 'list' 
                  ? 'bg-gray-900 text-white' 
                  : 'bg-white text-gray-600 hover:bg-gray-50'
              ]"
            >
              <List :size="16" />
            </button>
            <div class="w-px bg-gray-300"></div>
            <button
              @click="viewMode = 'grid'"
              :title="viewMode === 'grid' ? 'Grid view (active)' : 'Grid view'"
              :class="[
                'px-3 py-1.5 transition-colors flex items-center justify-center',
                viewMode === 'grid' 
                  ? 'bg-gray-900 text-white' 
                  : 'bg-white text-gray-600 hover:bg-gray-50'
              ]"
            >
              <Grid3x3 :size="16" />
            </button>
          </div>
        </div>
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
          <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-1 text-[10px] sm:text-[11px] text-gray-500 uppercase tracking-widest">
            <span class="text-left">Search and/or pick tags; all selected tags must be present.</span>
            <button v-if="selectedTags.length" type="button" @click="clearTags" class="text-blue-600 hover:text-blue-800 font-semibold whitespace-nowrap">Clear tags</button>
          </div>
        </div>
      </div>

      <UiSystemNotice
        v-model="statusMessage"
        :type="statusMessage?.type ?? 'success'"
        :title="statusMessage?.text"
      />

      <div v-if="filteredFiles.length === 0" class="py-12 text-center text-sm text-gray-500 italic">
        No files attached to this document.
      </div>

      <!-- List View -->
      <div v-else-if="viewMode === 'list'" class="space-y-0 divide-y divide-gray-200 border-t border-gray-200">
        <div 
          v-for="file in filteredFiles" 
          :key="file.file_name"
          :class="[
            'group py-3 px-2 sm:px-3 transition-colors hover:bg-gray-50',
            highlightActive === file.file_name
              ? 'bg-yellow-100 shadow-sm'
              : ''
          ]"
          :ref="el => setRowRef(file.file_name, el as HTMLElement | null)"
        >
          <div class="flex items-start justify-between gap-2">
            <div class="flex-1 min-w-0">
              <div class="flex items-center gap-2 mb-1 flex-wrap">
                <span class="text-xs sm:text-sm font-medium text-gray-900 break-all">{{ file.file_name }}</span>
                <span class="text-[10px] sm:text-[11px] font-mono text-gray-400 shrink-0">{{ formatSize(file.size) }}</span>
              </div>
              <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-4 gap-y-1 text-xs text-gray-600">
                <div v-if="file.description" class="wrap-break-word"><span class="text-gray-500">Desc:</span> {{ file.description }}</div>
                <div v-if="file.note" class="wrap-break-word"><span class="text-gray-500">Note:</span> {{ file.note }}</div>
                <div v-if="file.tags && file.tags.length > 0">
                  <span class="text-gray-500">Tags:</span> {{ formatTags(file.tags) }}
                </div>
              </div>
            </div>

            <div class="flex items-center gap-1 shrink-0">
              <button 
                @click="downloadFile(file.file_name)" 
                :title="'Download ' + file.file_name"
                class="p-1.5 sm:p-2 text-blue-600 hover:bg-blue-600 hover:text-white rounded transition-colors"
              >
                <Download :size="16" />
              </button>
              
              <template v-if="isAdmin">
                <button 
                  @click="openEditModal(file)"
                  :title="'Edit ' + file.file_name"
                  class="p-1.5 sm:p-2 text-gray-600 hover:bg-gray-800 hover:text-white rounded transition-colors"
                >
                  <Edit :size="16" />
                </button>
                <button 
                  @click="deleteFile(file.file_name)" 
                  :title="'Delete ' + file.file_name"
                  class="p-1.5 sm:p-2 text-red-600 hover:bg-red-600 hover:text-white rounded transition-colors"
                >
                  <Trash2 :size="16" />
                </button>
              </template>
            </div>
          </div>
        </div>
      </div>

      <!-- Grid View -->
      <div v-else class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-2 border-t border-gray-200 pt-3">
        <div 
          v-for="file in filteredFiles" 
          :key="file.file_name"
          :class="[
            'group relative rounded border transition-colors hover:shadow-md flex flex-col',
            highlightActive === file.file_name
              ? 'border-yellow-400 bg-yellow-50 shadow-md'
              : 'border-gray-200 bg-white hover:border-gray-300'
          ]"
          :ref="el => setRowRef(file.file_name, el as HTMLElement | null)"
        >
          <div class="p-2 space-y-1 flex-1">
            <div class="flex items-start justify-between gap-1">
              <div class="flex-1 min-w-0">
                <div class="h-10 sm:h-8 flex items-start">
                  <span class="text-xs font-medium text-gray-900 wrap-break-word line-clamp-2 leading-tight">{{ file.file_name }}</span>
                </div>
              </div>
              <div class="text-[10px] text-gray-400 shrink-0 font-mono whitespace-nowrap">{{ formatSize(file.size) }}</div>
            </div>
            
            <div class="h-10 sm:h-8 text-xs text-gray-600 overflow-hidden">
              <div v-if="file.description" class="line-clamp-2 wrap-break-word">{{ file.description }}</div>
              <div v-else class="text-gray-400 italic text-xs">No desc</div>
            </div>

            <div v-if="file.tags && file.tags.length > 0" class="flex flex-wrap gap-0.5">
              <span v-for="tag in (Array.isArray(file.tags) ? file.tags : [file.tags]).slice(0, 2)" :key="tag" class="inline-block text-[8px] bg-gray-100 text-gray-700 px-1 py-0.5 rounded">
                {{ tag }}
              </span>
              <span v-if="(Array.isArray(file.tags) ? file.tags : [file.tags]).length > 2" class="text-[8px] text-gray-500">+{{ (Array.isArray(file.tags) ? file.tags : [file.tags]).length - 2 }}</span>
            </div>
          </div>

          <div class="border-t border-gray-200 p-1 flex gap-0.5 justify-center">
            <button 
              @click="downloadFile(file.file_name)"
              :title="'Download ' + file.file_name"
              class="flex-1 h-8 flex items-center justify-center text-blue-600 hover:bg-blue-600 hover:text-white rounded transition-colors"
            >
              <Download :size="16" />
            </button>
            
            <template v-if="isAdmin">
              <button 
                @click="openEditModal(file)"
                :title="'Edit ' + file.file_name"
                class="flex-1 h-8 flex items-center justify-center text-gray-600 hover:bg-gray-800 hover:text-white rounded transition-colors"
              >
                <Edit :size="16" />
              </button>
              <button 
                @click="openDeleteFileConfirm(file.file_name)"
                :title="'Delete ' + file.file_name"
                class="flex-1 h-8 flex items-center justify-center text-red-600 hover:bg-red-600 hover:text-white rounded transition-colors"
              >
                <Trash2 :size="16" />
              </button>
            </template>
          </div>
        </div>
      </div>
    </UiCard>

    <!-- Edit Modal -->
    <div 
      v-if="editingFileName"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
      @click.self="closeEditModal"
    >
      <div class="bg-white rounded-lg shadow-lg w-full max-w-md max-h-[90vh] overflow-y-auto">
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-4 sm:p-6 border-b border-gray-200 sticky top-0 bg-white">
          <h2 class="text-sm sm:text-base font-bold text-gray-900 truncate pr-4">{{ editingFileName }}</h2>
          <button 
            @click="closeEditModal" 
            class="p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-colors shrink-0"
            title="Close"
          >
            <X :size="20" />
          </button>
        </div>

        <!-- Modal Tabs -->
        <div class="flex gap-0 border-b border-gray-300 bg-gray-50 px-4 sm:px-6">
          <button
            @click="editTab = 'metadata'"
            :class="[
              'flex-1 px-3 py-3 text-xs sm:text-sm font-bold transition-colors border-b-2',
              editTab === 'metadata'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900'
            ]"
          >
            Metadata
          </button>
          <button
            @click="editTab = 'update'"
            :class="[
              'flex-1 px-3 py-3 text-xs sm:text-sm font-bold transition-colors border-b-2',
              editTab === 'update'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900'
            ]"
          >
            Update File
          </button>
          <button
            @click="editTab = 'rename'"
            :class="[
              'flex-1 px-3 py-3 text-xs sm:text-sm font-bold transition-colors border-b-2',
              editTab === 'rename'
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-gray-600 hover:text-gray-900'
            ]"
          >
            Rename
          </button>
        </div>

        <!-- Modal Content -->
        <div class="p-4 sm:p-6 space-y-4">
          <!-- Metadata Tab -->
          <div v-if="editTab === 'metadata'" class="space-y-4">
            <div>
              <label class="block text-xs sm:text-sm font-bold text-gray-600 uppercase mb-2">Description</label>
              <input 
                v-model="editForm.description" 
                class="w-full text-sm p-2 border bg-white border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" 
              />
            </div>
            <div>
              <label class="block text-xs sm:text-sm font-bold text-gray-600 uppercase mb-2">Note</label>
              <textarea 
                v-model="editForm.note" 
                class="w-full text-sm p-2 border bg-white border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" 
                rows="3"
              ></textarea>
            </div>
            <div>
              <label class="block text-xs sm:text-sm font-bold text-gray-600 uppercase mb-2">Tags</label>
              <div class="space-y-2">
                <div v-for="(_, index) in editForm.tags" :key="index" class="flex gap-2">
                  <input 
                    v-model="editForm.tags[index]" 
                    placeholder="Enter tag" 
                    class="flex-1 text-sm p-2 border bg-white border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
                  />
                  <button 
                    @click="editForm.tags.splice(index, 1)" 
                    :title="'Remove tag'"
                    class="p-2 text-red-600 hover:bg-red-600 hover:text-white rounded transition-colors shrink-0"
                  >
                    <X :size="18" />
                  </button>
                </div>
                <button 
                  @click="editForm.tags.push('')" 
                  class="w-full px-3 py-2 text-xs sm:text-sm font-bold text-blue-600 hover:bg-blue-50 rounded transition-colors flex items-center justify-center gap-2"
                >
                  <Plus :size="18" />
                  <span>Add Tag</span>
                </button>
              </div>
            </div>
          </div>

          <!-- Update Tab -->
          <div v-if="editTab === 'update'" class="space-y-4">
            <div>
              <label class="block text-xs sm:text-sm font-bold text-gray-600 uppercase mb-2">Upload New Version</label>
              <input 
                type="file" 
                :id="`edit-file-input-modal-${editingFileName}`" 
                class="hidden" 
                @change="(e) => {
                  const file = (e.target as HTMLInputElement).files?.[0];
                  if (file && editingFileName) updateFile(editingFileName, file);
                }" 
              />
              <button 
                @click="triggerFileInput(`edit-file-input-modal-${editingFileName}`)"
                class="w-full px-4 py-3 border-2 border-dashed border-gray-300 rounded text-sm font-bold text-gray-600 hover:border-blue-400 hover:bg-blue-50 transition-colors"
              >
                Click to select new file
              </button>
            </div>
          </div>

          <!-- Rename Tab -->
          <div v-if="editTab === 'rename'" class="space-y-4">
            <div>
              <label class="block text-xs sm:text-sm font-bold text-gray-600 uppercase mb-2">Current Name</label>
              <div class="w-full text-sm p-2 bg-gray-100 border border-gray-300 rounded text-gray-700 font-mono break-all">
                {{ editingFileName }}
              </div>
            </div>
            <div>
              <label class="block text-xs sm:text-sm font-bold text-gray-600 uppercase mb-2">New Name</label>
              <input 
                v-model="editForm.new_name" 
                class="w-full text-sm p-2 border bg-white border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" 
              />
            </div>
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex justify-end gap-2 p-4 sm:p-6 border-t border-gray-200 bg-gray-50 sticky bottom-0">
          <button 
            @click="closeEditModal" 
            class="text-sm font-medium text-gray-600 hover:text-gray-900 px-4 py-2 rounded hover:bg-gray-100 transition-colors"
          >
            Cancel
          </button>
          <button 
            @click="handleModalSave" 
            class="text-sm font-bold bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700 transition-colors"
          >
            {{ editTab === 'metadata' ? 'Save Metadata' : editTab === 'rename' ? 'Rename' : 'Close' }}
          </button>
        </div>
      </div>
    </div>

    <!-- Delete File Modal -->
    <div 
      v-if="showDeleteFileConfirm"
      class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50"
      @click.self="showDeleteFileConfirm = false"
    >
      <div class="bg-white rounded-lg shadow-lg w-full max-w-md max-h-[90vh] overflow-y-auto">
        <!-- Modal Header -->
        <div class="flex items-center justify-between p-4 sm:p-6 border-b border-gray-200 sticky top-0 bg-white">
          <h2 class="text-sm sm:text-base font-bold text-gray-900">Delete File</h2>
          <button 
            @click="showDeleteFileConfirm = false" 
            class="p-1 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-colors shrink-0"
            title="Close"
          >
            <X :size="20" />
          </button>
        </div>

        <!-- Modal Content -->
        <div class="p-4 sm:p-6 space-y-4">
          <p class="text-gray-700 wrap-break-word mb-2">
            Are you sure you want to delete "<strong>{{ fileToDelete }}</strong>"?
          </p>
          <div class="bg-red-50 p-3 rounded border border-red-100">
            <p class="text-sm text-red-800 font-medium">This action is permanent.</p>
            <p class="text-sm text-red-700 mt-1 mb-2">
              To confirm, please type <span class="font-mono font-bold">"I want to delete it"</span> below:
            </p>
            <input 
              v-model="deleteFileConfirmInput" 
              type="text"
              placeholder="Type the confirmation phrase"
              @keyup.enter="canDeleteFile && fileToDelete && deleteFile(fileToDelete)"
              class="w-full bg-white border border-gray-300 text-sm p-2 rounded focus:ring-1 focus:ring-blue-500 outline-none"
            />
          </div>
        </div>

        <!-- Modal Footer -->
        <div class="flex justify-end gap-2 p-4 sm:p-6 border-t border-gray-200 bg-gray-50 sticky bottom-0">
          <button 
            @click="showDeleteFileConfirm = false" 
            class="text-sm font-medium text-gray-600 hover:text-gray-900 px-4 py-2 rounded hover:bg-gray-100 transition-colors"
          >
            Cancel
          </button>
          <button 
            @click="fileToDelete && deleteFile(fileToDelete)" 
            :disabled="!canDeleteFile"
            class="text-sm font-bold bg-red-600 text-white px-4 py-2 rounded hover:bg-red-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Delete Permanently
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, watch } from 'vue';
import { Download, Edit, Trash2, X, Plus, List, Grid3x3 } from 'lucide-vue-next';
import { UiCard } from '@/core/ui';
import { DocumentService } from '@/modules/documents/services/documentService';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';

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
const viewMode = ref<'list' | 'grid'>('grid');
const highlightActive = ref('');
const pendingHighlight = ref('');
const highlightTimer = ref<ReturnType<typeof setTimeout> | null>(null);
const rowRefs = ref<Record<string, HTMLElement | null>>({});
const isHighlighting = ref(false);

const showDeleteFileConfirm = ref(false);
const fileToDelete = ref<string | null>(null);
const deleteFileConfirmInput = ref('');
const REQUIRED_PHRASE_FILE = 'i want to delete it';
const canDeleteFile = computed(() => deleteFileConfirmInput.value.toLowerCase() === REQUIRED_PHRASE_FILE);

const searchQuery = ref('');
const selectedTags = ref<(string | number)[]>([]);
const editingFileName = ref<string | null>(null);
const editTab = ref<'metadata' | 'rename' | 'update'>('metadata');
const editForm = ref<{ description: string; note: string; tags: string[]; new_name?: string }>({ description: '', note: '', tags: [] });

function triggerFileInput(inputId: string) {
  const input = document.getElementById(inputId) as HTMLInputElement;
  if (input) input.click();
}

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

function formatTags(tags: any): string {
  if (Array.isArray(tags)) {
    return tags.join(', ');
  }
  return String(tags);
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

function openEditModal(file: any) {
  editingFileName.value = file.file_name;
  const tags = Array.isArray(file.tags) ? [...file.tags] : [];
  editTab.value = 'metadata';
  editForm.value = { 
    description: file.description || '', 
    note: file.note || '',
    tags,
    new_name: file.file_name
  };
}

function closeEditModal() {
  editingFileName.value = null;
}

function handleModalSave() {
  if (editTab.value === 'metadata' && editingFileName.value) {
    saveMetadata(editingFileName.value);
  } else if (editTab.value === 'rename' && editingFileName.value) {
    renameFile(editingFileName.value);
  }
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
      description: editForm.value.description,
      note: editForm.value.note,
      tags: editForm.value.tags.filter(t => t.trim())
    };
    
    await DocumentService.updateFileMetadata(props.document.uuid, fileName, metadata);
    editingFileName.value = null;
    emit('refresh');
    statusMessage.value = { type: 'success', text: 'metadata updated' };
  } catch (err: any) {
    statusMessage.value = { type: 'error', text: 'update failed' };
  }
}

async function renameFile(oldFileName: string) {
  if (!editForm.value.new_name || editForm.value.new_name === oldFileName) {
    statusMessage.value = { type: 'error', text: 'new filename must be different' };
    return;
  }
  try {
    await DocumentService.renameFile(props.document.uuid, oldFileName, { new_name: editForm.value.new_name });
    editingFileName.value = null;
    emit('refresh');
    statusMessage.value = { type: 'success', text: `renamed to ${editForm.value.new_name}` };
  } catch (err: any) {
    statusMessage.value = { type: 'error', text: 'rename failed' };
  }
}

async function updateFile(fileName: string, file: File) {
  try {
    // Delete old file and upload new one with same name
    await DocumentService.deleteFile(props.document.uuid, fileName);
    await DocumentService.uploadFile(props.document.uuid, file);
    editingFileName.value = null;
    emit('refresh');
    statusMessage.value = { type: 'success', text: `${fileName} updated successfully` };
  } catch (err: any) {
    statusMessage.value = { type: 'error', text: 'update failed' };
  }
}

function openDeleteFileConfirm(fileName: string) {
  fileToDelete.value = fileName;
  showDeleteFileConfirm.value = true;
}

async function deleteFile(fileName: string) {
  if (!fileToDelete.value) return;
  try {
    await DocumentService.deleteFile(props.document.uuid, fileName);
    showDeleteFileConfirm.value = false;
    fileToDelete.value = null;
    emit('refresh');
    statusMessage.value = { type: 'success', text: `${fileName} deleted successfully` };
  } catch (err: any) {
    statusMessage.value = { type: 'error', text: 'DELETE_FAILED' };
    showDeleteFileConfirm.value = false;
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
    
    // Highlight the uploaded file
    pendingHighlight.value = file.name;
    await nextTick();
    const exists = filteredFiles.value.some((f: any) => f.file_name === file.name);
    if (exists) {
      highlightAndScroll(file.name);
      pendingHighlight.value = '';
    }
    
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
    if (exists) {
      highlightAndScroll(fileName);
      pendingHighlight.value = '';
    }
  }
);

watch(filteredFiles, () => {
  if (isHighlighting.value || !pendingHighlight.value) return;
  const target = pendingHighlight.value;
  if (filteredFiles.value.some((f: any) => f.file_name === target)) {
    highlightAndScroll(target);
    pendingHighlight.value = '';
  }
});

watch(showDeleteFileConfirm, (isOpen) => {
  if (!isOpen) deleteFileConfirmInput.value = '';
});

function formatSize(bytes?: number) {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
</script>