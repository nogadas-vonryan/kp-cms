<template>
  <UiCard>
    <div 
      class="mb-2 border-2 border-dashed rounded p-6 transition-colors duration-200 text-center"
      :class="[
        isDragging ? 'border-blue-600 bg-blue-50' : 'border-gray-300 bg-gray-50',
        isUploading ? 'opacity-50 pointer-events-none' : 'cursor-pointer',
        statusMessage?.type === 'error' ? 'border-red-600' : ''
      ]"
      @dragover.prevent="isDragging = true"
      @dragleave.prevent="isDragging = false"
      @drop.prevent="handleDrop"
      @click="fileInput?.click()"
    >
      <input type="file" ref="fileInput" class="hidden" @change="handleFileSelect" />
      
      <div class="space-y-2">
        <p class="text-sm font-medium text-gray-900">
          {{ isUploading ? 'Processing System Upload...' : 'Click to upload or drag and drop' }}
        </p>
        <p class="text-[10px] text-gray-500 font-mono uppercase tracking-wider">
          All System Formats Supported (Docs, Media, Archives) • Max 32MB
        </p>
      </div>
    </div>

    <UiSystemNotice 
      v-model="statusMessage"
      type="success"
      label="SYSTEM_IO"
      :title="statusMessage?.text"
      dismissible
    />

    <hr class="border-gray-200 mb-6" />

    <div v-if="!document.files || document.files.length === 0" class="text-center text-gray-600 py-8 italic text-sm">
      No files attached to system record {{ document.uuid.split('-')[0] }}...
    </div>
    <div v-else class="space-y-2">
      <div
        v-for="file in document.files"
        :key="file.file_name"
        class="flex items-center justify-between p-3 border border-gray-200 rounded bg-white hover:border-gray-400 transition-colors"
      >
        <div class="min-w-0">
          <p class="font-mono text-sm font-medium text-gray-900 truncate">{{ file.file_name }}</p>
          <p class="text-xs text-gray-600">
            {{ file.type?.split('/')[1]?.toUpperCase() || 'FILE' }} • {{ formatSize(file.size) }} • {{ formatDate(file.created_at) }}
          </p>
        </div>
        <div class="flex gap-2 ml-4">
          <button class="px-3 py-1 text-xs font-medium border border-gray-300 bg-white text-gray-900 hover:bg-gray-50 rounded">
            Download
          </button>
        </div>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import type { Document } from '@/types';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';
import { DocumentService } from '@/modules/documents/services/documentService'; //

const props = defineProps<{
  document: Document;
}>();

const emit = defineEmits(['refresh']);

const isDragging = ref(false);
const isUploading = ref(false);
const fileInput = ref<HTMLInputElement | null>(null);
const statusMessage = ref<{ type: 'success' | 'error', text: string } | null>(null);

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

function formatDate(dateStr?: string) {
  if (!dateStr) return 'Pending';
  return new Date(dateStr).toLocaleDateString();
}

function formatSize(bytes?: number) {
  if (!bytes) return '0 B';
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
</script>