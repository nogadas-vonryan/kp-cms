<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-gray-900">Admin Dashboard</h1>

    <!-- System Actions -->
    <UiCard>
      <div class="space-y-4">
        <h2 class="text-lg font-semibold text-gray-900">System Actions</h2>
        <div class="flex gap-2">
          <UiButton @click="handleReload" :loading="reloading">
            Reload Documents
          </UiButton>
        </div>
        
        <UiAlert v-if="reloadMessage" :type="reloadSuccess ? 'success' : 'error'">
          {{ reloadMessage }}
        </UiAlert>
      </div>
    </UiCard>

    <!-- Conflicts -->
    <UiCard>
      <div class="space-y-4">
        <div class="flex items-center justify-between">
          <h2 class="text-lg font-semibold text-gray-900">
            Sync Conflicts
            <span v-if="conflicts && conflicts.length > 0" class="ml-2 text-red-600">({{ conflicts && conflicts.length }})</span>
          </h2>
          <UiButton @click="loadConflicts" :loading="loadingConflicts">
            Refresh
          </UiButton>
        </div>

        <UiAlert v-if="conflictsError" type="error">
          {{ conflictsError }}
        </UiAlert>

        <div v-else-if="!conflicts || conflicts.length === 0" class="text-center py-8 text-gray-600">
          No conflicts detected
        </div>

        <div v-else class="space-y-3">
          <UiSystemNotice
            v-for="(conflict, idx) in conflicts || []"
            :key="idx"
            type="warning"
            :label="conflict.type"
            :modelValue="true" 
          >
            <template #title>
              <span class="text-sm font-medium text-gray-900">{{ conflict.path }}</span>
            </template>
            
            <p class="text-sm text-gray-700">{{ conflict.message }}</p>
          </UiSystemNotice>
        </div>
      </div>
    </UiCard>

    <!-- System Info -->
    <UiCard>
      <div class="space-y-3">
        <h2 class="text-lg font-semibold text-gray-900">System Information</h2>
        <div class="space-y-2 text-sm">
          <div class="flex justify-between">
            <span class="text-gray-600">Role:</span>
            <span class="font-medium text-gray-900">{{ authStore.role }}</span>
          </div>
          <div class="flex justify-between">
            <span class="text-gray-600">Username:</span>
            <span class="font-medium text-gray-900">{{ authStore.user?.id || 'N/A' }}</span>
          </div>
        </div>
      </div>
    </UiCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import type { SyncIssue } from '@/types';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';

const authStore = useAuthStore();

const conflicts = ref<SyncIssue[]>([]);
const loadingConflicts = ref(false);
const conflictsError = ref('');

const reloading = ref(false);
const reloadMessage = ref('');
const reloadSuccess = ref(false);

async function loadConflicts() {
  loadingConflicts.value = true;
  conflictsError.value = '';
  
  try {
    const response = await DocumentService.listConflicts();
    conflicts.value = response.data;
  } catch (err: any) {
    conflictsError.value = err.response?.data?.error || 'Failed to load conflicts';
  } finally {
    loadingConflicts.value = false;
  }
}

async function handleReload() {
  reloading.value = true;
  reloadMessage.value = '';
  
  try {
    const response = await DocumentService.reload();
    reloadSuccess.value = response.data?.status === 'success';
    const respConflicts = Array.isArray(response.data?.conflicts) ? response.data.conflicts : [];
    reloadMessage.value = `Reload ${response.data?.status}. ${respConflicts.length} conflicts found.`;

    // Refresh conflicts list
    if (respConflicts.length > 0) {
      conflicts.value = respConflicts;
    } else {
      await loadConflicts();
    }
  } catch (err: any) {
    reloadSuccess.value = false;
    reloadMessage.value = err.response?.data?.error || 'Reload failed';
  } finally {
    reloading.value = false;
  }
}

onMounted(() => {
  loadConflicts();
});
</script>
