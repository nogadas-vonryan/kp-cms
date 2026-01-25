<template>
  <div class="space-y-4">
    <UiCard>
      <div class="space-y-4">
        <UiAlert v-if="error" type="error" class="mb-4">{{ error }}</UiAlert>

        <!-- Title -->
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Title</label>
          <div v-if="!isEditing" class="text-lg font-semibold text-gray-900 truncate block max-w-4xl">
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

        <!-- Custom Fields Section -->
        <div v-if="Object.keys(document?.fields || {}).length > 0 || isEditing" class="pt-2 border-t border-gray-200">
          <div class="flex items-center justify-between mb-3">
            <h3 class="font-semibold text-gray-900">Custom Fields</h3>
            <UiButton
              v-if="isEditing"
              @click="showAddFieldModal = true"
              variant="secondary"
              class="text-sm"
            >
              + Add Field
            </UiButton>
          </div>

          <div v-if="Object.keys(editForm.fields).length > 0" class="space-y-4">
            <div v-for="([key, value]) in sortedFields" :key="key" :class="[
              'rounded p-3',
              isEditing ? 'border border-gray-200' : 'border-l-4 border-gray-300 bg-gray-50'
            ]">
              <!-- Field Header -->
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900">{{ formatLabel(key) }}</span>
                <div v-if="isEditing" class="flex gap-2">
                  <button v-if="!Array.isArray(value)" @click="convertToArray(key)" class="text-xs px-2 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded">Make Array</button>
                  <button v-else @click="convertToText(key)" class="text-xs px-2 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded">Make Text</button>
                  <button @click="deleteFieldConfirm(key)" class="text-xs px-2 py-1 bg-red-50 text-red-600 hover:bg-red-100 rounded">Delete</button>
                </div>
              </div>
              
              <!-- Read-only View -->
              <div v-if="!isEditing">
                 <div v-if="!Array.isArray(value)" class="text-sm text-gray-700 whitespace-pre-wrap">
                    {{ getFieldDisplay(key, value) }}
                 </div>
                 <div v-else class="space-y-1">
                    <div v-for="(item, index) in value" :key="index" class="text-sm text-gray-700 whitespace-pre-wrap">- {{ item || '—' }}</div>
                 </div>
              </div>

              <!-- Editable View -->
              <div v-else>
                 <div v-if="!Array.isArray(value)" class="space-y-2">
                    <UiSelect v-if="getFieldOptions(key)" :model-value="value" @update:model-value="(v: any) => editForm.fields[key] = v" :options="getFieldOptions(key)!" />
                    <UiTextarea v-else :model-value="value" @update:model-value="(v: any) => editForm.fields[key] = v" :rows="getTextareaRows(value)" />
                 </div>
                 <div v-else class="space-y-2">
                    <div v-for="(item, index) in value" :key="index" class="flex gap-2 items-start">
                      <UiTextarea :model-value="item" @update:model-value="(v) => (editForm.fields[key] as any[])[index] = v" :rows="getTextareaRows(item)" class="flex-1" />
                      <button @click="removeArrayItem(key, index)" class="text-red-600 hover:text-red-800 text-sm px-2 mt-2">Remove</button>
                    </div>
                    <button @click="addArrayItem(key)" class="text-xs px-3 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded mt-2">+ Add Item</button>
                 </div>
              </div>
            </div>
          </div>

          <div v-else-if="!isEditing" class="text-sm text-gray-600 text-center py-4">
            No custom fields
          </div>
        </div>

        <!-- Sticky Footer for Save/Cancel -->
        <div v-if="isEditing" class="sticky bottom-0 z-10 -mx-6 -mb-6 mt-4 bg-white/90 backdrop-blur-sm border-t border-gray-200 p-4 flex gap-2 shadow-[0_-4px_6px_-1px_rgba(0,0,0,0.05)]">
          <UiButton @click="saveChanges" :loading="saving" :disabled="!hasChanges || !editForm.title.trim()">Save Changes</UiButton>
          <UiButton @click="emit('cancel')" variant="secondary">Cancel</UiButton>
        </div>
      </div>
    </UiCard>

    <!-- Add Field Modal -->
    <UiModal v-model:open="showAddFieldModal" title="Add Custom Field">
      <form @submit.prevent="handleAddField" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Field Name *</label>
          <UiInput v-model="newField.name" placeholder="e.g., status, category" required />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Field Type</label>
          <UiSelect :model-value="newField.isArray ? '1' : '0'" @update:model-value="(v) => (newField.isArray = v === '1')" :options="[{ value: '0', label: 'Text' }, { value: '1', label: 'Array' }]" />
        </div>
        <div v-if="newField.isArray">
          <label class="block text-sm font-medium text-gray-700 mb-1">Initial Items (comma-separated)</label>
          <UiInput v-model="newField.initialValue" placeholder="item1, item2, item3" />
        </div>
        <div v-else>
          <label class="block text-sm font-medium text-gray-700 mb-1">Initial Value</label>
          <UiTextarea v-model="newField.initialValue" :rows="3" />
        </div>
      </form>
      <template #footer>
        <UiButton @click="showAddFieldModal = false">Cancel</UiButton>
        <UiButton @click="handleAddField">Add Field</UiButton>
      </template>
    </UiModal>

    <!-- Delete Field Confirmation -->
    <UiModal v-model:open="showDeleteFieldConfirm" title="Delete Field">
      <p class="text-gray-700">Are you sure you want to delete "<strong>{{ formatLabel(fieldToDelete) }}</strong>"?</p>
      <template #footer>
        <UiButton @click="showDeleteFieldConfirm = false">Cancel</UiButton>
        <UiButton @click="confirmDeleteField" variant="danger">Delete</UiButton>
      </template>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import type { Document } from '@/types';
import { DocumentService } from '@/modules/documents/services/documentService';
import { useDocumentFields } from '@/modules/documents/composables/useDocumentFields';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiTextarea from '@/core/ui/components/UiTextarea.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiSelect from '@/core/ui/components/UiSelect.vue';

const props = defineProps<{
  document: Document;
  isEditing: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:document', doc: Document): void;
  (e: 'cancel'): void;
  (e: 'saved'): void;
}>();

const { getFieldOptions, formatLabel, toSnakeCase, getTextareaRows, getSortedFields } = useDocumentFields();

const saving = ref(false);
const error = ref('');
const showAddFieldModal = ref(false);
const showDeleteFieldConfirm = ref(false);
const fieldToDelete = ref('');
const newField = ref({ name: '', isArray: false, initialValue: '' });

// Form State
const editForm = ref({
  title: '',
  code: '',
  folder_name: '',
  fields: {} as Record<string, any>
});

// Initialize form when document changes or edit mode starts
watch(() => props.document, (newDoc) => {
  if (newDoc) {
    editForm.value = {
      title: newDoc.title || '',
      code: newDoc.code || '',
      folder_name: newDoc.folder_name || '',
      fields: JSON.parse(JSON.stringify(newDoc.fields || {}))
    };
  }
}, { immediate: true });

// Also reset form when cancelling edit (reverting changes)
watch(() => props.isEditing, (isEditing) => {
  if (!isEditing && props.document) {
     editForm.value = {
      title: props.document.title || '',
      code: props.document.code || '',
      folder_name: props.document.folder_name || '',
      fields: JSON.parse(JSON.stringify(props.document.fields || {}))
    };
  }
});

const sortedFields = computed(() => getSortedFields(editForm.value.fields));

const hasChanges = computed(() => {
  if (!props.document) return false;
  return (
    editForm.value.title !== props.document.title ||
    JSON.stringify(editForm.value.fields) !== JSON.stringify(props.document.fields)
  );
});

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString();
}

function getFieldDisplay(key: string, value: any): string {
  const options = getFieldOptions(key);
  const found = options?.find((opt: { value: any; label: string }) => opt.value === value);
  if (found?.label) return found.label;
  const str = value?.toString ? value.toString() : '';
  return formatLabel(str) || '—';
}

// --- Field Manipulation Logic ---

function convertToArray(key: string) {
  const current = editForm.value.fields[key];
  if (!Array.isArray(current)) {
    editForm.value.fields[key] = current ? [current] : [''];
  }
}

function convertToText(key: string) {
  const current = editForm.value.fields[key];
  if (Array.isArray(current)) {
    const joined = current.map((item) => (item ?? '').toString()).filter(i => i.length > 0).join('\n');
    editForm.value.fields[key] = joined;
  }
}

function addArrayItem(key: string) {
  if (Array.isArray(editForm.value.fields[key])) {
    editForm.value.fields[key].push('');
  }
}

function removeArrayItem(key: string, index: number) {
  if (Array.isArray(editForm.value.fields[key])) {
    editForm.value.fields[key].splice(index, 1);
  }
}

function deleteFieldConfirm(key: string) {
  fieldToDelete.value = key;
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
  const fieldName = toSnakeCase(newField.value.name);
  
  if (editForm.value.fields.hasOwnProperty(fieldName)) {
    alert('Field already exists');
    return;
  }

  if (newField.value.isArray) {
    const items = newField.value.initialValue.split(',').map(i => i.trim()).filter(i => i.length > 0);
    editForm.value.fields[fieldName] = items.length > 0 ? items : [''];
  } else {
    editForm.value.fields[fieldName] = newField.value.initialValue;
  }
  
  newField.value = { name: '', isArray: false, initialValue: '' };
  showAddFieldModal.value = false;
}

async function saveChanges() {
  if (!props.document) return;
  saving.value = true;
  error.value = '';
  
  try {
    await DocumentService.update(props.document.uuid, {
      title: editForm.value.title,
      code: editForm.value.code,
      fields: editForm.value.fields
    });
    // Assuming service returns the updated document structure
    // If not, we trigger a reload from parent via 'saved' event
    emit('saved');
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to save changes';
  } finally {
    saving.value = false;
  }
}
</script>