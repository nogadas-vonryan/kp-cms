<template>
  <div class="space-y-4">
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-900">Import</h1>
    </div>

    <!-- Tab Navigation -->
    <div class="border-b border-gray-200 overflow-x-auto -mx-4 sm:mx-0 px-4 sm:px-0">
      <div class="flex gap-4 sm:gap-8 min-w-min">
        <button
          v-for="tab in tabs"
          :key="tab.id"
          @click="activeTab = tab.id"
          :class="[
            'px-2 sm:px-4 py-2 font-medium border-b-2 transition-colors text-sm sm:text-base whitespace-nowrap',
            activeTab === tab.id
              ? 'border-blue-600 text-blue-600'
              : 'border-transparent text-gray-600 hover:text-gray-900'
          ]"
        >
          {{ tab.label }}
        </button>
      </div>
    </div>

    <!-- Tab Content -->
    <DocumentImportTab v-if="activeTab === 'documents'" />
    <InhabitantImportTab v-if="activeTab === 'inhabitants'" />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import DocumentImportTab from '../components/DocumentImportTab.vue';
import InhabitantImportTab from '../components/InhabitantImportTab.vue';

const activeTab = ref('documents');

const tabs = [
  { id: 'documents', label: 'Documents' },
  { id: 'inhabitants', label: 'Inhabitants' },
];
</script>