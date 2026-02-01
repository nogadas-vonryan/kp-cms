<template>
  <div class="space-y-4">
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-900">System Backups</h1>
      <div class="flex flex-col items-end gap-2 w-full sm:w-auto">
        <UiButton 
          variant="secondary" 
          @click="handleCreateBackup"
          :disabled="exporting || isSystemUninitialized || restoring"
          class="flex-1 sm:flex-none flex items-center justify-center gap-2"
        >
          <Save :size="18" />
          <span>{{ exporting ? 'Generating...' : 'Create New Backup' }}</span>
        </UiButton>
        
        <div v-if="exporting" class="w-full sm:w-64 space-y-1">
          <div class="flex justify-between text-[10px] font-bold uppercase text-blue-600">
            <span>Compressing Files</span>
            <span>{{ Math.round(backupProgress) }}%</span>
          </div>
          <div class="w-full bg-gray-200 rounded-full h-1.5 overflow-hidden">
            <div 
              class="bg-blue-600 h-full transition-all duration-700 ease-out shadow-[0_0_8px_rgba(37,99,235,0.4)]"
              :style="{ width: `${backupProgress}%` }"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <UiAlert v-if="isSystemUninitialized" type="error" class="bg-red-50 border-red-200">
      <div class="flex gap-3">
        <Folder :size="20" class="shrink-0 text-red-600" />
        <div>
          <p class="font-bold text-red-800">Backup Directory Not Found</p>
          <p class="text-xs text-red-700">The server backup directory (./backup) does not exist.</p>
        </div>
      </div>
    </UiAlert>

    <UiCard class="relative min-h-64">
      <div v-if="loading" class="absolute inset-0 z-20 flex items-center justify-center bg-white/60 backdrop-blur-[1px]">
        <span class="text-sm font-medium text-gray-600">Loading backups...</span>
      </div>
      
      <div class="w-full overflow-x-auto">
        <UiSystemNotice v-if="error && !isSystemUninitialized" type="error" :modelValue="true" class="m-4">
          <template #title><span class="px-2 text-sm text-gray-700">{{ error }}</span></template>
        </UiSystemNotice>

        <div v-if="backups.length > 0 && !isSystemUninitialized" class="table-wrapper">
          <table class="w-full text-sm">
            <thead class="border-b border-gray-200 bg-gray-50">
              <tr>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">File Name</th>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Size</th>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Created At</th>
                <th class="text-right px-4 py-3 font-semibold text-gray-900">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="file in backups" :key="file.file_name" class="hover:bg-gray-50 transition-colors">
                <td class="px-4 py-3 font-mono text-xs text-blue-600">{{ file.file_name }}</td>
                <td class="px-4 py-3 text-gray-600">{{ formatFileSize(file.size) }}</td>
                <td class="px-4 py-3 text-gray-600">{{ formatDate(file.created_at) }}</td>
                <td class="px-4 py-3 text-right">
                  <div class="flex justify-end gap-1">
                    <button 
                      @click="downloadExisting(file.file_name)"
                      class="p-2 text-gray-500 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                      title="Download"
                    >
                      <Download :size="16" />
                    </button>
                    <button 
                      v-if="isAdmin"
                      @click="prepareRestore(file.file_name)"
                      :disabled="restoring"
                      class="p-2 text-gray-500 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors disabled:opacity-50"
                      title="Restore this version"
                    >
                      <RotateCcw :size="16" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-else-if="!loading || isSystemUninitialized" class="p-48 text-center text-gray-500">
          <Database :size="48" class="mx-auto mb-3 text-gray-300" />
          <div v-if="isSystemUninitialized">Feature Disabled: Directory Missing</div>
          <div v-else>No backup history found</div>
        </div>
      </div>
    </UiCard>

    <UiModal v-model:open="showRestoreModal" title="Restore System Backup" :prevent-close="restoring">
      <div class="space-y-4">
        <div class="bg-blue-50 p-4 rounded-lg border border-blue-200 flex items-center gap-3">
          <FileArchive :size="24" class="text-blue-600" />
          <div>
            <p class="text-xs font-semibold text-blue-800 uppercase tracking-wider">Target File</p>
            <p class="text-sm font-mono text-blue-700">{{ fileToRestore }}</p>
          </div>
        </div>

        <div v-if="restoring" class="py-4 space-y-3">
          <div class="flex justify-between items-end">
            <span class="text-sm font-medium text-gray-700 animate-pulse">Extracting and Restoring...</span>
            <span class="text-lg font-bold text-blue-600">{{ Math.round(restoreProgress) }}%</span>
          </div>
          <div class="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
            <div 
              class="bg-blue-600 h-full transition-all duration-700 ease-out"
              :style="{ width: `${restoreProgress}%` }"
            ></div>
          </div>
          <p class="text-[10px] text-gray-400 text-center uppercase tracking-widest">Do not refresh or close the browser</p>
        </div>

        <template v-else>
          <div class="space-y-2">
            <label class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Restore Mode</label>
            <div class="flex gap-2 p-1 bg-gray-100 rounded-lg">
              <button 
                v-for="mode in (['merge', 'overwrite'] as const)" 
                :key="mode"
                @click="restoreMode = mode"
                :class="[
                  'flex-1 py-1.5 text-xs font-medium rounded-md capitalize transition-all',
                  restoreMode === mode ? 'bg-white shadow text-blue-600' : 'text-gray-500 hover:text-gray-700'
                ]"
              >
                {{ mode }}
              </button>
            </div>
          </div>

          <UiAlert v-if="restoreError" type="error" class="text-xs">
            {{ restoreError }}
          </UiAlert>

          <div v-if="restoreMode === 'overwrite'" class="bg-red-50 p-4 rounded-lg border border-red-200 space-y-3">
            <div class="flex gap-2 text-red-800">
              <AlertTriangle :size="18" class="shrink-0" />
              <p class="text-xs font-bold uppercase tracking-wide">Destructive Action Confirmation</p>
            </div>
            <p class="text-xs text-red-700">
              This will <strong>trash</strong> your current data and replace it. To confirm, type:
              <span class="block font-mono font-bold mt-1 bg-white p-1 border border-red-100 rounded text-center">
                {{ OVERWRITE_PHRASE }}
              </span>
            </p>
            <UiInput 
              v-model="confirmationInput" 
              placeholder="Type the confirmation phrase"
              class="bg-white"
            />
          </div>
        </template>
      </div>

      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <UiButton :disabled="restoring" @click="showRestoreModal = false">Cancel</UiButton>
          <UiButton 
            v-if="!restoring"
            @click="handleRestore" 
            :disabled="!canRestore" 
            :variant="restoreMode === 'overwrite' ? 'danger' : 'secondary'"
            class="flex items-center gap-2"
          >
            <RotateCcw :size="16" />
            <span>Confirm Restore</span>
          </UiButton>
        </div>
      </template>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { 
  Download, Database, AlertTriangle, Save, Folder, RotateCcw, FileArchive 
} from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import { extractErrorMessage } from '@/core/api';
import type { BackupFile } from '@/types';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';

const authStore = useAuthStore();
const isAdmin = computed(() => authStore.role === 'RoleAdmin');

// List/Backup State
const backups = ref<BackupFile[]>([]);
const loading = ref(false);
const exporting = ref(false);
const backupProgress = ref(0);
const error = ref('');
const isSystemUninitialized = ref(false);

// Restoration State
const restoring = ref(false);
const restoreProgress = ref(0);
const restoreError = ref('');
const showRestoreModal = ref(false);
const fileToRestore = ref<string | null>(null);
const restoreMode = ref<'merge' | 'overwrite'>('merge');
const confirmationInput = ref('');
const OVERWRITE_PHRASE = 'I want to delete and replace everything';

async function fetchBackups() {
  if (!authStore.user) return;
  loading.value = true;
  error.value = '';
  try {
    const response = await DocumentService.listBackups();
    backups.value = response.data.backups || [];
  } catch (err: any) {
    const errMsg = extractErrorMessage(err) || '';
    if (errMsg.includes('no such file or directory')) isSystemUninitialized.value = true;
    else error.value = errMsg;
  } finally {
    loading.value = false;
  }
}

async function downloadExisting(fileName: string) {
  try {
    const blob = await DocumentService.downloadBackup(fileName);
    const url = window.URL.createObjectURL(new Blob([blob]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', fileName);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  } catch (err: any) {
    alert('Download failed: ' + extractErrorMessage(err));
  }
}

function prepareRestore(fileName: string) {
  fileToRestore.value = fileName;
  restoreProgress.value = 0;
  restoreError.value = '';
  showRestoreModal.value = true;
}

const canRestore = computed(() => {
  if (!fileToRestore.value || restoring.value) return false;
  if (restoreMode.value === 'overwrite') {
    return confirmationInput.value.toLowerCase() === OVERWRITE_PHRASE;
  }
  return true;
});

/**
 * Initiates the backup job
 */
async function handleCreateBackup() {
  if (exporting.value) return;
  exporting.value = true;
  // Start at 1% immediately so the bar is visible even before polling
  backupProgress.value = 1; 
  
  try {
    const response = await DocumentService.createBackup();
    const jobId = response.job_id; 
    pollJobStatus(jobId, 'backup');
  } catch (err: any) {
    exporting.value = false;
    error.value = 'Backup failed: ' + extractErrorMessage(err);
  }
}

/**
 * Initiates the restore job
 */
async function handleRestore() {
  if (!canRestore.value || !fileToRestore.value) return;
  
  restoring.value = true;
  restoreError.value = '';
  // Start at 1% immediately so the bar is visible
  restoreProgress.value = 1;

  try {
    const response = await DocumentService.restoreBackup(fileToRestore.value, restoreMode.value);
    pollJobStatus(response.data.job_id, 'restore');
  } catch (err: any) {
    restoring.value = false;
    restoreError.value = extractErrorMessage(err) || 'Failed to start restore';
  }
}

/**
 * Enhanced polling with minimum display time for animations
 */
function pollJobStatus(jobId: string, type: 'backup' | 'restore') {
  let retryCount = 0;

  // Reduced interval to 500ms for more frequent updates
  const intervalId = window.setInterval(async () => {
    try {
      const { data } = await DocumentService.getJobStatus(jobId);
      retryCount = 0;

      if (type === 'backup') {
        backupProgress.value = data.progress;
      } else {
        restoreProgress.value = data.progress;
      }

      if (data.status === 'completed') {
        clearInterval(intervalId);
        
        // Force to 100% to ensure the CSS transition triggers
        if (type === 'backup') backupProgress.value = 100;
        else restoreProgress.value = 100;

        // VISUAL DELAY: Give the CSS transition 800ms to complete before hiding the bar
        setTimeout(() => {
          finalizeJob(type, true);
        }, 800);

      } else if (data.status === 'failed') {
        clearInterval(intervalId);
        finalizeJob(type, false, data.error);
      }
    } catch (err: any) {
      if (err.statusCode === 404 && retryCount < 3) {
        retryCount++;
        return; 
      }
      clearInterval(intervalId);
      finalizeJob(type, false, "Job tracking failed or expired");
    }
  }, 500); 
}

function finalizeJob(type: 'backup' | 'restore', success: boolean, errMsg?: string) {
  if (type === 'backup') {
    exporting.value = false;
    if (success) {
      fetchBackups();
    } else {
      backupProgress.value = 0;
      error.value = errMsg || 'Backup process failed';
    }
  } else {
    restoring.value = false;
    if (success) {
      setTimeout(() => {
        showRestoreModal.value = false;
        fetchBackups();
        alert('System successfully restored.');
      }, 500);
    } else {
      restoreProgress.value = 0;
      restoreError.value = errMsg || 'Restore process failed';
    }
  }
}

function formatDate(dateStr: string) { return new Date(dateStr).toLocaleString(); }
function formatFileSize(bytes: number) {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

onMounted(fetchBackups);
</script>