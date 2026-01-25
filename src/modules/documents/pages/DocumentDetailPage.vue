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
        <UiButton
          v-if="isAdmin && !editMode"
          @click="enterEditMode"
          variant="secondary"
        >
          Edit
        </UiButton>
        <UiButton
          v-if="isAdmin && editMode"
          @click="cancelEdit"
          variant="secondary"
        >
          Cancel
        </UiButton>
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
            <!-- Title -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
              <div v-if="!editMode" class="text-lg font-semibold text-gray-900 truncate block max-w-4xl">
                {{ document.title }}
              </div>
              <UiInput
                v-else
                v-model="editForm.title"
                required
                placeholder="Document title"
              />
            </div>

            <!-- Document Info -->
            <div class="grid grid-cols-2 gap-4 pt-2 border-t border-gray-200">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Code</label>
                <div class="text-sm font-mono text-gray-600">{{ document.code }}</div>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Folder</label>
                <div class="text-sm text-gray-600">{{ document.folder_name || 'N/A' }}</div>
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
            <div v-if="Object.keys(document.fields).length > 0 || editMode" class="pt-2 border-t border-gray-200">
              <div class="flex items-center justify-between mb-3">
                <h3 class="font-semibold text-gray-900">Custom Fields</h3>
                <UiButton
                  v-if="editMode"
                  @click="showAddFieldModal = true"
                  variant="secondary"
                  class="text-sm"
                >
                  + Add Field
                </UiButton>
              </div>

              <div v-if="Object.keys(editForm.fields).length > 0" class="space-y-4">
                <div v-for="(value, key) in editForm.fields" :key="key" :class="[
                  'rounded p-3',
                  editMode ? 'border border-gray-200' : 'border-l-4 border-gray-300 bg-gray-50'
                ]">
                  <!-- Field Header -->
                  <div class="flex items-center justify-between mb-2">
                    <span class="font-medium text-gray-900">{{ key }}</span>
                    <div v-if="editMode" class="flex gap-2">
                      <button
                        v-if="!Array.isArray(value)"
                        @click="convertToArray(key)"
                        class="text-xs px-2 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded"
                        title="Convert to array field"
                      >
                        Make Array
                      </button>
                      <button
                        v-else
                        @click="convertToText(key)"
                        class="text-xs px-2 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded"
                        title="Convert to text field"
                      >
                        Make Text
                      </button>
                      <button
                        @click="deleteFieldConfirm(key)"
                        class="text-xs px-2 py-1 bg-red-50 text-red-600 hover:bg-red-100 rounded"
                      >
                        Delete
                      </button>
                    </div>
                  </div>
                  
                  <!-- Text Field (Read-only) -->
                  <div v-if="!editMode && !Array.isArray(value)" class="text-sm text-gray-700 whitespace-pre-wrap">
                    {{ value || '—' }}
                  </div>

                  <!-- Text Field (Editable) -->
                  <div v-else-if="editMode && !Array.isArray(value)" class="space-y-2">
                    <UiSelect
                      v-if="getFieldOptions(key)"
                      :model-value="value"
                      @update:model-value="(v: any) => editForm.fields[key] = v"
                      :options="getFieldOptions(key)"
                    />
                    
                    <UiTextarea
                      v-else
                      :model-value="value"
                      @update:model-value="(v: any) => editForm.fields[key] = v"
                      placeholder="Field value"
                      :rows="getTextareaRows(value)"
                    />
                  </div>

                  <!-- Array Field (Read-only) -->
                  <div v-else-if="!editMode && Array.isArray(value)" class="space-y-1">
                    <div v-for="(item, index) in value" :key="index" class="text-sm text-gray-700 whitespace-pre-wrap">
                      - {{ item || '—' }}
                    </div>
                  </div>

                  <!-- Array Field (Editable) -->
                  <div v-else-if="editMode && Array.isArray(value)" class="space-y-2">
                    <div
                      v-for="(item, index) in value"
                      :key="index"
                      class="flex gap-2 items-start"
                    >
                      <UiTextarea
                        :model-value="item"
                        @update:model-value="(v) => (editForm.fields[key] as any[])[index] = v"
                        placeholder="Item value"
                        :rows="getTextareaRows(item)"
                        class="flex-1"
                      />
                      <button
                        @click="removeArrayItem(key, index)"
                        class="text-red-600 hover:text-red-800 text-sm px-2 mt-2"
                      >
                        Remove
                      </button>
                    </div>
                    <button
                      @click="addArrayItem(key)"
                      class="text-xs px-3 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded mt-2"
                    >
                      + Add Item
                    </button>
                  </div>
                </div>
              </div>

              <div v-else-if="!editMode" class="text-sm text-gray-600 text-center py-4">
                No custom fields
              </div>
            </div>

            <!-- Edit Form Footer -->
            <div v-if="editMode" class="pt-4 border-t border-gray-200 flex gap-2">
              <UiButton 
                @click="saveChanges" 
                :loading="saving" 
                :disabled="!hasChanges || !editForm.title.trim()"
              >
                Save Changes
              </UiButton>
              <UiButton @click="cancelEdit" variant="secondary">
                Cancel
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
            :model-value="newField.isArray ? '1' : '0'"
            @update:model-value="(v) => (newField.isArray = v === '1')"
            :options="[
              { value: '0', label: 'Text' },
              { value: '1', label: 'Array' }
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
          <UiTextarea
            v-model="newField.initialValue"
            placeholder="Field value"
            :rows="3"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Quick Add</label>
          <div class="flex gap-2 mb-4">
            <button 
              type="button"
              @click="newField.name = 'nature'; newField.isArray = false"
              class="text-xs px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded"
            >+ Nature</button>
            <button 
              type="button"
              @click="newField.name = 'status'; newField.isArray = false"
              class="text-xs px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded"
            >+ Status</button>
          </div>
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
import { ref, computed, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import { usePluginStore } from '@/core/plugins/pluginRegistry';
import type { Document, PluginContext } from '@/types';
import PluginHost from '@/core/plugins/pluginHost';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiTextarea from '@/core/ui/components/UiTextarea.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiSelect from '@/core/ui/components/UiSelect.vue';

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
const editMode = ref(false);

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

// Define predefined options for specific fields
const FIELD_OPTIONS = {
  nature: [
    { value: 'civil', label: 'Civil' },
    { value: 'criminal', label: 'Criminal' }
  ],
  status: [
    { value: 'case_filed', label: 'Case Filed' },
    { value: 'arbitration', label: 'Arbitration' },
    { value: 'mediation', label: 'Mediation' },
    { value: 'conciliation', label: 'Conciliation' },
    { value: 'repudiation', label: 'Repudiation' }
  ]
};

function getFieldOptions(key: string) {
  const opts = FIELD_OPTIONS[key as keyof typeof FIELD_OPTIONS] as { value: string; label: string }[] | undefined;
  return opts ? [...opts] : undefined;
}

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

function getTextareaRows(value: any): number {
  if (!value) return 1;
  const text = value.toString();
  // Use 2 rows if text contains newlines or is longer than 60 characters
  if (text.includes('\n') || text.length > 60) {
    return 2;
  }
  return 1;
}

function convertToArray(fieldName: string) {
  const current = editForm.value.fields[fieldName];
  if (!Array.isArray(current)) {
    editForm.value.fields[fieldName] = current ? [current] : [''];
  }
}

function convertToText(fieldName: string) {
  const current = editForm.value.fields[fieldName];
  if (Array.isArray(current)) {
    const joined = current
      .map((item) => (item ?? '').toString())
      .filter((item) => item.length > 0)
      .join('\n');
    editForm.value.fields[fieldName] = joined;
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
}

function enterEditMode() {
  editMode.value = true;
}

function cancelEdit() {
  editMode.value = false;
  // Reset form to original document state
  if (document.value) {
    editForm.value = {
      title: document.value.title,
      code: document.value.code,
      folder_name: document.value.folder_name,
      fields: JSON.parse(JSON.stringify(document.value.fields))
    };
  }
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
      // Deep clone to avoid sharing array/object references with document.value
      fields: JSON.parse(JSON.stringify(response.data.fields))
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
    editMode.value = false;
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to save changes';
  } finally {
    saving.value = false;
  }
}

// Delete confirmation
const deleteConfirmationInput = ref('');
const REQUIRED_PHRASE = 'i want to delete it';

const canDelete = computed(() => {
  return deleteConfirmationInput.value.toLowerCase() === REQUIRED_PHRASE;
});

watch(showDeleteConfirm, (isOpen) => {
  if (!isOpen) {
    deleteConfirmationInput.value = '';
  }
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
