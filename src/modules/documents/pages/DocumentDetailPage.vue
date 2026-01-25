<template>
  <div class="space-y-4">
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-sm text-gray-600">
      <router-link to="/documents" class="hover:text-gray-900">Documents</router-link>
      <span>/</span>
      <span class="text-gray-900 font-medium">{{ document?.title || 'Loading...' }}</span>
    </div>

    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-3xl font-bold text-gray-900">{{ document?.title }}</h1>
        <p class="text-sm text-gray-600 mt-1">
          Code: <span class="font-mono">{{ document?.code }}</span>
        </p>
      </div>
      <div v-if="isAdmin" class="flex gap-2">
        <UiButton @click="showDeleteConfirm = true" variant="danger">
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

    <!-- Document Content -->
    <div v-if="document" class="space-y-4">
      <!-- Tabs -->
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

      <!-- Details Tab -->
      <div v-if="activeTab === 'details'" class="space-y-4">
        <UiCard>
          <div class="space-y-4">
            <!-- Title (editable) -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
              <div v-if="!isAdmin" class="text-sm">{{ document.title }}</div>
              <UiInput
                v-else
                v-model="editForm.title"
                required
                placeholder="Document title"
              />
            </div>

            <!-- Document Info -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Code</label>
                <div class="text-sm font-mono">{{ document.code }}</div>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Folder</label>
                <div class="text-sm">{{ document.folder_name || 'N/A' }}</div>
              </div>
            </div>

            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Created</label>
                <div class="text-sm text-gray-600">{{ formatDate(document.created_at) }}</div>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Updated</label>
                <div class="text-sm text-gray-600">{{ formatDate(document.updated_at) }}</div>
              </div>
            </div>

            <!-- Custom Fields -->
            <div v-if="Object.keys(document.fields).length > 0 || isAdmin">
              <div class="flex items-center justify-between mb-3">
                <h3 class="font-semibold text-gray-900">Custom Fields</h3>
                <UiButton
                  v-if="isAdmin"
                  @click="showAddFieldModal = true"
                  variant="secondary"
                  class="text-sm"
                >
                  + Add Field
                </UiButton>
              </div>

              <div v-if="Object.keys(editForm.fields).length > 0" class="space-y-4">
                <div v-for="(value, key) in editForm.fields" :key="key" class="border border-gray-200 rounded p-3">
                  <!-- Field Header -->
                  <div class="flex items-center justify-between mb-2">
                    <span class="font-medium text-gray-900">{{ key }}</span>
                    <div v-if="isAdmin" class="flex gap-2">
                      <button
                        v-if="!Array.isArray(value)"
                        @click="convertToArray(key)"
                        class="text-xs px-2 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded"
                        title="Convert to array field"
                      >
                        Make Array
                      </button>
                      <button
                        @click="deleteFieldConfirm(key)"
                        class="text-xs px-2 py-1 bg-red-50 text-red-600 hover:bg-red-100 rounded"
                      >
                        Delete
                      </button>
                    </div>
                  </div>

                  <!-- Text Field -->
                  <div v-if="!Array.isArray(value)" class="space-y-2">
                    <UiInput
                      v-if="isAdmin"
                      :value="value"
                      @update:model-value="(v) => editForm.fields[key] = v"
                      placeholder="Field value"
                    />
                    <div v-else class="text-sm">{{ value }}</div>
                  </div>

                  <!-- Array Field -->
                  <div v-else class="space-y-2">
                    <div
                      v-for="(item, index) in value"
                      :key="index"
                      class="flex gap-2 items-center"
                    >
                      <UiInput
                        v-if="isAdmin"
                        :model-value="item"
                        @update:model-value="(v) => (editForm.fields[key] as any[])[index] = v"
                        placeholder="Item value"
                        class="flex-1"
                      />
                      <div v-else class="flex-1 text-sm">{{ item }}</div>
                      <button
                        v-if="isAdmin"
                        @click="removeArrayItem(key, index)"
                        class="text-red-600 hover:text-red-800 text-sm px-2"
                      >
                        Remove
                      </button>
                    </div>
                    <button
                      v-if="isAdmin"
                      @click="addArrayItem(key)"
                      class="text-xs px-3 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded mt-2"
                    >
                      + Add Item
                    </button>
                  </div>
                </div>
              </div>

              <div v-else class="text-sm text-gray-600 text-center py-4">
                No custom fields
              </div>
            </div>

            <!-- Edit Form -->
            <div v-if="isAdmin" class="pt-4 border-t border-gray-200">
              <UiButton 
                @click="saveChanges" 
                :loading="saving" 
                :disabled="!hasChanges || !editForm.title.trim()"
              >
                Save Changes
              </UiButton>
            </div>
          </div>
        </UiCard>
      </div>

      <!-- Files Tab -->
      <div v-if="activeTab === 'files'" class="space-y-4">
        <UiCard>
          <div v-if="document.files.length === 0" class="text-center text-gray-600 py-8">
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
              <button class="text-blue-600 hover:text-blue-800 text-sm">
                Download
              </button>
            </div>
          </div>
        </UiCard>
      </div>

      <!-- Plugin Tabs -->
      <div v-for="plugin in pluginsWithLocation" :key="plugin.id" v-show="activeTab === plugin.id">
        <PluginHost
          :plugins="[plugin]"
          :location="`${plugin.id}Tab` as any"
          :context="pluginContext"
        />
      </div>
    </div>

    <!-- Delete Confirmation Modal -->
    <UiModal v-model:open="showDeleteConfirm" title="Delete Document">
      <p class="text-gray-700">
        Are you sure you want to delete "<strong>{{ document?.title }}</strong>"?
      </p>
      <p class="text-sm text-gray-600 mt-2">This action cannot be undone.</p>
      <template #footer>
        <UiButton @click="showDeleteConfirm = false">Cancel</UiButton>
        <UiButton @click="handleDelete" :loading="deleting" variant="danger">Delete</UiButton>
      </template>
    </UiModal>

    <!-- Add Field Modal -->
    <UiModal v-model:open="showAddFieldModal" title="Add Custom Field">
      <form @submit.prevent="handleAddField" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Field Name *</label>
          <UiInput
            v-model="newField.name"
            placeholder="e.g., status, category"
            required
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Field Type</label>
          <UiSelect
            v-model="newField.isArray"
            :options="[
              { value: false, label: 'Text' },
              { value: true, label: 'Array' }
            ]"
          />
        </div>
        <div v-if="newField.isArray">
          <label class="block text-sm font-medium text-gray-700 mb-1">Initial Items (comma-separated, optional)</label>
          <UiInput
            v-model="newField.initialValue"
            placeholder="item1, item2, item3"
          />
        </div>
        <div v-else>
          <label class="block text-sm font-medium text-gray-700 mb-1">Initial Value (optional)</label>
          <UiInput
            v-model="newField.initialValue"
            placeholder="Field value"
          />
        </div>
      </form>
      <template #footer>
        <UiButton @click="showAddFieldModal = false">Cancel</UiButton>
        <UiButton @click="handleAddField">Add Field</UiButton>
      </template>
    </UiModal>

    <!-- Delete Field Confirmation Modal -->
    <UiModal v-model:open="showDeleteFieldConfirm" title="Delete Field">
      <p class="text-gray-700">
        Are you sure you want to delete the field "<strong>{{ fieldToDelete }}</strong>"?
      </p>
      <p class="text-sm text-gray-600 mt-2">This action cannot be undone.</p>
      <template #footer>
        <UiButton @click="showDeleteFieldConfirm = false">Cancel</UiButton>
        <UiButton @click="confirmDeleteField" variant="danger">Delete</UiButton>
      </template>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import { usePluginStore } from '@/core/plugins/pluginRegistry';
import type { Document, PluginContext } from '@/types';
import PluginHost from '@/core/plugins/pluginHost';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';

const router = useRouter();
const authStore = useAuthStore();
const pluginStore = usePluginStore();

const isAdmin = computed(() => authStore.role === 'RoleAdmin');

// State
const document = ref<Document | null>(null);
const loading = ref(false);
const error = ref('');
const saving = ref(false);
const deleting = ref(false);
const activeTab = ref('details');
const showDeleteConfirm = ref(false);

const tabs = computed(() => [
  { id: 'details', label: 'Details' },
  { id: 'files', label: 'Files' },
  ...pluginsWithLocation.value.map(p => ({
    id: p.id,
    label: p.name
  }))
]);

const editForm = ref({
  title: '',
  code: '',
  folder_name: '',
  fields: {} as Record<string, any>
});

const showAddFieldModal = ref(false);
const showDeleteFieldConfirm = ref(false);
const fieldToDelete = ref('');

const newField = ref({
  name: '',
  isArray: false,
  initialValue: ''
});

const pluginsWithLocation = computed(() => {
  return pluginStore.plugins.filter(p => {
    const tabLocation = `${p.id}Tab` as keyof typeof p.locations;
    return p.locations[tabLocation];
  });
});

const pluginContext = computed<PluginContext>(() => ({
  document: document.value || undefined,
  isAdmin: isAdmin.value,
  events: {
    onDocumentUpdate: (doc: Document) => {
      document.value = doc;
    },
    onFileAdd: () => {
      loadDocument();
    },
    onError: (err: string) => {
      error.value = err;
    }
  }
}));

const hasChanges = computed(() => {
  if (!document.value) return false;
  return (
    editForm.value.title !== document.value.title ||
    editForm.value.code !== document.value.code ||
    editForm.value.folder_name !== document.value.folder_name ||
    JSON.stringify(editForm.value.fields) !== JSON.stringify(document.value.fields)
  );
});

function convertToArray(fieldName: string) {
  const current = editForm.value.fields[fieldName];
  if (!Array.isArray(current)) {
    editForm.value.fields[fieldName] = current ? [current] : [''];
  }
}

function addArrayItem(fieldName: string) {
  const field = editForm.value.fields[fieldName];
  if (Array.isArray(field)) {
    field.push('');
  }
}

function removeArrayItem(fieldName: string, index: number) {
  const field = editForm.value.fields[fieldName];
  if (Array.isArray(field)) {
    field.splice(index, 1);
  }
}

function deleteFieldConfirm(fieldName: string) {
  fieldToDelete.value = fieldName;
  showDeleteFieldConfirm.value = true;
}

function confirmDeleteField() {
  if (fieldToDelete.value) {
    delete editForm.value.fields[fieldToDelete.value];
    fieldToDelete.value = '';
    showDeleteFieldConfirm.value = false;
  }
}

function handleAddField() {
  if (!newField.value.name.trim()) return;

  const fieldName = newField.value.name.trim();
  if (editForm.value.fields.hasOwnProperty(fieldName)) {
    alert('Field already exists');
    return;
  }

  if (newField.value.isArray) {
    // Parse comma-separated values
    const items = newField.value.initialValue
      .split(',')
      .map(item => item.trim())
      .filter(item => item.length > 0);
    editForm.value.fields[fieldName] = items.length > 0 ? items : [''];
  } else {
    editForm.value.fields[fieldName] = newField.value.initialValue;
  }

  // Reset form
  newField.value = { name: '', isArray: false, initialValue: '' };
  showAddFieldModal.value = false;
};

async function loadDocument() {
  loading.value = true;
  error.value = '';
  try {
    const documentId = router.currentRoute.value.params.documentId as string;
    const response = await DocumentService.getById(documentId);
    document.value = response.data;
    
    // Initialize edit form
    editForm.value = {
      title: response.data.title,
      code: response.data.code,
      folder_name: response.data.folder_name,
      fields: { ...response.data.fields }
    };
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to load document';
    console.error('Error loading document:', err);
  } finally {
    loading.value = false;
  }
}

async function saveChanges() {
  if (!document.value) return;
  
  saving.value = true;
  error.value = '';
  try {
    await DocumentService.update(document.value.uuid, {
      title: editForm.value.title,
      code: editForm.value.code,
      fields: editForm.value.fields
    });
    await loadDocument();
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to save changes';
  } finally {
    saving.value = false;
  }
}

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

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString();
}

function formatSize(bytes: number) {
  if (bytes < 1024) return bytes + ' B';
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
}

onMounted(() => {
  loadDocument();
});
</script>
