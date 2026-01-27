<template>
  <div class="space-y-4">
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-sm text-gray-600">
      <router-link to="/documents" class="hover:text-gray-900">Documents</router-link>
      <span>/</span>
      <span class="text-gray-900 font-medium truncate block max-w-4xl">{{ document?.title || 'Loading...' }}</span>
    </div>

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900 truncate block max-w-4xl">{{ document?.title }}</h1>
        <p class="text-sm text-gray-600 mt-1">
          Code: <span class="font-mono">{{ document?.code }}</span>
        </p>
      </div>
      <div class="flex gap-2">
        <!-- Edit Buttons (Only show if in Details tab and Admin) -->
        <template v-if="activeTab === 'details'">
            <UiButton
            v-if="isAdmin && !editMode"
            @click="editMode = true"
            variant="secondary"
            >
            Edit
            </UiButton>
            <UiButton
            v-if="isAdmin && editMode"
            @click="editMode = false"
            variant="secondary"
            >
            Cancel
            </UiButton>
        </template>
        
        <UiButton v-if="isAdmin" @click="showDeleteConfirm = true" variant="danger">
          Delete
        </UiButton>
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
      <div class="border-b border-gray-200">
        <div class="flex gap-8">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="activeTab = tab.id"
            :class="[
              'px-4 py-2 font-medium border-b-2 transition-colors',
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
      <p class="text-gray-700 truncate block max-w-xl mb-2">
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
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import { usePluginStore } from '@/core/plugins/pluginRegistry';
import type { Document, PluginContext } from '@/types';
import PluginHost from '@/core/plugins/pluginHost';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiInput from '@/core/ui/components/UiInput.vue';

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
const activeTab = ref('details');
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
        return { document: document.value };
    }
    return {};
});

const pluginContext = computed<PluginContext>(() => ({
  document: document.value || undefined,
  isAdmin: isAdmin.value,
  events: {
    onDocumentUpdate: (doc: Document) => { document.value = doc; },
    onFileAdd: () => { loadDocument(); },
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

onMounted(() => {
  loadDocument();
});
</script>