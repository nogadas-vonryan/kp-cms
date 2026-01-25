<template>
  <UiCard>
    <div v-if="!document.files || document.files.length === 0" class="text-center text-gray-600 py-8">
      No files attached
    </div>
    <div v-else class="space-y-3">
      <div
        v-for="file in document.files"
        :key="file.file_name"
        class="flex items-center justify-between p-3 border border-gray-200 rounded"
      >
        <div>
          <p class="font-medium text-gray-900">{{ file.file_name }}</p>
          <p class="text-sm text-gray-600">
            {{ file.type }} • {{ formatSize(file.size) }} • {{ formatDate(file.created_at) }}
          </p>
        </div>
        <div class="flex gap-2">
            <!-- Placeholder for future actions like Delete/Preview -->
            <button class="text-blue-600 hover:text-blue-800 text-sm">
                Download
            </button>
        </div>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import type { Document } from '@/types';
import UiCard from '@/core/ui/components/UiCard.vue';

defineProps<{
  document: Document;
}>();

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString();
}

function formatSize(bytes: number) {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}
</script>