<template>
  <div class="space-y-4">
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-xs sm:text-sm text-gray-600 min-w-0">
      <router-link to="/documents" class="hover:text-gray-900 truncate">Documents</router-link>
      <span class="shrink-0">/</span>
      <span class="text-gray-900 font-medium wrap-break-word">{{ document?.title || 'Loading...' }}</span>
    </div>

    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div class="min-w-0">
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 wrap-break-word">{{ document?.title }}</h1>
        <p class="text-xs sm:text-sm text-gray-600 mt-1">
          Code: <span class="font-mono">{{ document?.code }}</span>
        </p>
      </div>
      <div class="flex gap-2 shrink-0 ml-auto">
        <!-- Edit Buttons (Only show if in Details tab and Admin) -->
        <template v-if="activeTab === 'details'">
            <button
            v-if="isAdmin && !editMode"
            @click="editMode = true"
            title="Edit"
            class="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded border border-gray-300 transition-colors"
            >
            <PencilIcon :size="20" :stroke-width="1.4" />
            </button>
            <button
            v-if="isAdmin && editMode"
            @click="editMode = false"
            title="Cancel"
            class="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded border border-gray-300 transition-colors"
            >
            <XIcon :size="20" :stroke-width="1.4" />
            </button>
        </template>

        <button 
          v-if="isAdmin" 
          @click="handleReload"
          title="Reload"
          :disabled="reloading"
          class="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded border border-gray-300 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          <RefreshCwIcon :size="20" :stroke-width="1.4" :class="reloading && 'animate-spin'" />
        </button>
        
        <button v-if="isAdmin" @click="showDeleteConfirm = true" title="Delete" class="p-2 text-red-600 hover:text-red-900 hover:bg-red-50 rounded border border-red-300 transition-colors">
          <Trash2Icon :size="20" :stroke-width="1.4" />
        </button>
      </div>
    </div>

    <!-- Loading & Error States -->
    <div v-if="loading" class="p-8 text-center text-gray-600">
      Loading document...
    </div>

    <UiAlert v-if="error" type="error" class="mb-4">
      {{ error }}
    </UiAlert>

    <!-- Main Content -->
    <div v-if="document" class="space-y-4">
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

      <!-- Tab Content Area -->
      <KeepAlive>
        <component 
          :is="currentTabComponent" 
          v-bind="currentTabProps"
          @saved="onDocumentSaved"
          @cancel="editMode = false"
          @refresh="loadDocument"
          :is-admin="isAdmin"
        />
      </KeepAlive>
      
      <!-- Plugin Tabs (Rendered separately to maintain PluginHost context) -->
      <div v-for="plugin in pluginsWithLocation" :key="plugin.id" v-show="activeTab === plugin.id">
        <PluginHost
          :plugins="[plugin]"
          :location="`${plugin.id}Tab` as any"
          :context="pluginContext"
        />
      </div>
    </div>

    <!-- Global Delete Modal -->
    <UiModal v-model:open="showDeleteConfirm" title="Delete Document">
      <p class="text-gray-700 wrap-break-word mb-2">
        Are you sure you want to delete "<strong>{{ document?.title }}</strong>"?
      </p>
      <div class="bg-red-50 p-3 rounded border border-red-100">
        <p class="text-sm text-red-800 font-medium">This action is permanent.</p>
        <p class="text-sm text-red-700 mt-1 mb-2">
            To confirm, please type <span class="font-mono font-bold">"I want to delete it"</span> below:
        </p>
        <UiInput 
            v-model="deleteConfirmationInput" 
            placeholder="Type the confirmation phrase"
            @keyup.enter="canDelete && handleDelete()"
        />
      </div>
      <template #footer>
        <UiButton @click="showDeleteConfirm = false" variant="secondary">Cancel</UiButton>
        <UiButton 
          @click="handleDelete" 
          :loading="deleting" 
          variant="danger"
          :disabled="!canDelete" 
        >
          Delete Permanently
        </UiButton>
      </template>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { PencilIcon, XIcon, RefreshCwIcon, Trash2Icon } from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import { usePluginStore } from '@/core/plugins/pluginRegistry';
import type { Document, PluginContext } from '@/types';
import PluginHost from '@/core/plugins/pluginHost';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiButton from '@/core/ui/components/UiButton.vue';

// Sub-components
import DocumentDetailsTab from '@/modules/documents/components/DocumentDetailsTab.vue';
import DocumentFilesTab from '@/modules/documents/components/DocumentFilesTab.vue';

const router = useRouter();
const authStore = useAuthStore();
const pluginStore = usePluginStore();

const isAdmin = computed(() => authStore.role === 'RoleAdmin');

// State
const document = ref<Document | null>(null);
const loading = ref(false);
const error = ref('');
const deleting = ref(false);
const reloading = ref(false);
const activeTab = ref('details');
const highlightFileName = ref('');
const showDeleteConfirm = ref(false);
const editMode = ref(false);

const pluginsWithLocation = computed(() => {
  return pluginStore.plugins.filter(p => {
    const tabLocation = `${p.id}Tab` as keyof typeof p.locations;
    return p.locations[tabLocation];
  });
});

const tabs = computed(() => [
  { id: 'details', label: 'Details' },
  { id: 'files', label: 'Files' },
  ...pluginsWithLocation.value.map(p => ({
    id: p.id,
    label: p.name
  }))
]);

// Dynamic Component Resolution
const currentTabComponent = computed(() => {
    if (activeTab.value === 'details') return DocumentDetailsTab;
    if (activeTab.value === 'files') return DocumentFilesTab;
    return 'div'; // Fallback for plugins (handled via v-show below)
});

const currentTabProps = computed(() => {
    if (activeTab.value === 'details') {
        return { document: document.value, isEditing: editMode.value };
    }
    if (activeTab.value === 'files') {
		return { document: document.value, highlightFileName: highlightFileName.value };
    }
    return {};
});

const pluginContext = computed<PluginContext>(() => ({
  document: document.value || undefined,
  isAdmin: isAdmin.value,
  events: {
    onDocumentUpdate: (doc: Document) => { document.value = doc; },
    onFileAdd: (fileName?: string) => {
      highlightFileName.value = fileName || '';
      activeTab.value = 'files';
      loadDocument();
    },
    onError: (err: string) => { error.value = err; }
  }
}));

async function loadDocument() {
  loading.value = true;
  error.value = '';
  try {
    const documentId = router.currentRoute.value.params.documentId as string;
    const response = await DocumentService.getById(documentId);
    document.value = response.data || {};
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to load document';
  } finally {
    loading.value = false;
  }
}

function onDocumentSaved() {
    editMode.value = false;
    loadDocument(); // Refresh data to show updated fields
}

// Delete Logic
const deleteConfirmationInput = ref('');
const REQUIRED_PHRASE = 'i want to delete it';
const canDelete = computed(() => deleteConfirmationInput.value.toLowerCase() === REQUIRED_PHRASE);

watch(showDeleteConfirm, (isOpen) => {
  if (!isOpen) deleteConfirmationInput.value = '';
});

async function handleDelete() {
  if (!document.value) return;
  deleting.value = true;
  try {
    await DocumentService.remove(document.value.uuid);
    router.push('/documents');
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to delete document';
    deleting.value = false;
  }
}

async function handleReload() {
  if (!document.value) return;
  reloading.value = true;
  error.value = '';
  try {
    // Use the document's code or folder_name as the folderName parameter
    const folderName = (document.value as any).folder_name || document.value.code;
    if (!folderName) {
      throw new Error('Unable to determine folder name for reload');
    }
    await DocumentService.reloadDocument(folderName);
    await loadDocument();
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to reload document';
  } finally {
    reloading.value = false;
  }
}

onMounted(() => {
  if (!authStore.user) return;
  
  loadDocument();
});
</script>