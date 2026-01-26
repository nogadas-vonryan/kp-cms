<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-gray-900">Documents</h1>
      <UiButton v-if="isAdmin" @click="showCreateModal = true">
        Create Document
      </UiButton>
    </div>

    <!-- Search & Filters -->
    <UiCard>
      <form @submit.prevent="performSearch()">
        <div class="space-y-3">
          <div class="flex gap-2">
            <UiInput
              v-model="uuidSearchQuery"
              placeholder="Search by code..."
              class="flex-1"
            />
            <UiInput
              v-model="titleSearchQuery"
              placeholder="Filter by title..."
              class="flex-1"
            />
            <UiInput
              v-model="filters.folder_name"
              placeholder="Filter by folder..."
              class="flex-1"
            />
            <UiSelect
              v-model="filters.sort_by"
              :options="sortOptions"
              placeholder="Sort by..."
              class="flex-1"
            />
            <UiButton type="submit">Search</UiButton>
            <UiButton type="button" @click="resetSearch" variant="secondary">Reset</UiButton>
          </div>
          
          <details class="group">
            <summary class="text-sm text-blue-600 hover:text-blue-700 cursor-pointer font-medium list-none flex items-center gap-1 select-none">
              <span class="group-open:rotate-90 transition-transform">▶</span>
              Advanced Filters
            </summary>
            
            <div class="pt-3 space-y-3 border-t border-gray-100 mt-2">
              <div class="grid grid-cols-2 gap-3">
                <UiInput
                  v-model="filters.date_from"
                  type="date"
                  placeholder="From date"
                />
                <UiInput
                  v-model="filters.date_to"
                  type="date"
                  placeholder="To date"
                />
              </div>
              <div class="grid grid-cols-2 gap-3">
                <UiComboBox 
                  v-model="filters.field_key" 
                  placeholder="Select field..." 
                  :options="fieldOptions" 
                />

                <UiComboBox 
                  v-model="filters.field_value" 
                  :placeholder="filters.field_key ? `Select ${filters.field_key}...` : 'Select value...'" 
                  :options="dynamicValueOptions"
                  :disabled="!filters.field_key" 
                />
              </div>
            </div>
          </details>
        </div>
      </form>
    </UiCard>

    <!-- Documents Table -->
    <UiCard :padded="false" class="relative min-h-125">
      <div 
        v-if="loading" 
        class="absolute inset-0 z-20 flex items-center justify-center bg-white/60 backdrop-blur-[1px]"
      >
        <div class="flex flex-col items-center gap-2">
          <span class="text-sm font-medium text-gray-600">Loading documents...</span>
        </div>
      </div>
      
      <div class="w-full">
        <UiSystemNotice 
          v-if="error" 
          type="error" 
          label="System_Error"
          :modelValue="true"
          class="m-4"
        >
          <template #title>
            <span class="px-2 text-sm text-gray-700 leading-relaxed">
              {{ error }}
            </span>
          </template>
        </UiSystemNotice>

        <template v-if="documents.length > 0">
          <UiTable :columns="columns" :rows="documents">
            <template #cell:code="{ value }">
              <span class="font-mono text-sm">{{ value }}</span>
            </template>
            <template #cell:title="{ value }">
              <span class="text-gray-600 truncate block max-w-xl">{{ value }}</span>
            </template>
            <template #cell:status="{ row }">
              <span class="text-sm text-gray-600 truncate block max-w-50">{{ ((row as unknown) as Document).fields?.status || 'none' }}</span>
            </template>
            <template #cell:folder_name="{ value }">
              <span class="text-sm text-gray-600 truncate block max-w-50">{{ value }}</span>
            </template>
            <template #cell:files="{ row }">
              <span class="text-sm text-gray-600">{{ ((row as unknown) as Document).files.length }} file(s)</span>
            </template>
            <template #cell:created_at="{ value }">
              <span class="text-sm text-gray-600">{{ formatDate(value as string) }}</span>
            </template>
            <template #cell:actions="{ row }">
              <div class="flex gap-2">
                <router-link
                  :to="`/documents/${(row as unknown as Document).uuid}`"
                  class="text-blue-600 hover:text-blue-800 text-sm"
                >
                  View
                </router-link>
              </div>
            </template>
          </UiTable>
        </template>

        <div v-else-if="!loading" class="p-8 text-center text-gray-600">
          No documents found
        </div>
      </div>
    </UiCard>

    <!-- Pagination -->
    <div v-if="documents.length > 0" class="flex justify-between items-center">
      <span class="text-sm text-gray-600">
        Showing {{ offset + 1 }}-{{ Math.min(offset + limit, offset + documents.length) }}
      </span>
      <div class="flex gap-2">
        <UiButton @click="prevPage" :disabled="offset === 0">
          Previous
        </UiButton>
        <UiButton @click="nextPage" :disabled="documents.length < limit">
          Next
        </UiButton>
      </div>
    </div>

    <!-- Create Modal -->
    <UiModal v-model:open="showCreateModal" title="Create Document">
      <form @submit.prevent="handleCreate" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Title <span class="text-red-500">*</span></label>
          <UiInput v-model="form.title" required />
        </div>

        <div class="pt-4 border-t border-gray-200">
          <h3 class="font-semibold text-gray-900 mb-3">Document Details</h3>
          
          <div class="space-y-4">
            
            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Nature</span>
              </div>
              <UiSelect 
                v-model="form.fields.nature" 
                :options="[
                  { label: 'Civil', value: 'civil' }, 
                  { label: 'Criminal', value: 'criminal' } 
                ]"
                placeholder="Select nature" 
              />
            </div>

            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Complaint</span>
              </div>
              <UiTextarea v-model="form.fields.complaint" :rows="3" />
            </div>

            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Complainants</span>
              </div>
              <div class="space-y-2">
                <div v-for="(_, index) in form.fields.complainants" :key="index" class="flex gap-2 items-start">
                  <UiInput v-model="form.fields.complainants![index]" class="flex-1" placeholder="Name" />
                  <button 
                    type="button" 
                    @click="form.fields.complainants?.splice(index, 1)" 
                    class="text-red-600 hover:text-red-800 text-sm px-2 mt-2"
                  >
                    Remove
                  </button>
                </div>
                <button
                  type="button"
                  @click="form.fields.complainants?.push('')"
                  class="text-xs px-3 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded mt-2"
                >
                  + Add Item
                </button>
              </div>
            </div>

            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Respondents</span>
              </div>
              <div class="space-y-2">
                <div v-for="(_, index) in form.fields.respondents" :key="index" class="flex gap-2 items-start">
                  <UiInput v-model="form.fields.respondents![index]" class="flex-1" placeholder="Name" />
                  <button 
                    type="button" 
                    @click="form.fields.respondents?.splice(index, 1)" 
                    class="text-red-600 hover:text-red-800 text-sm px-2 mt-2"
                  >
                    Remove
                  </button>
                </div>
                <button
                  type="button"
                  @click="form.fields.respondents?.push('')"
                  class="text-xs px-3 py-1 bg-blue-50 text-blue-600 hover:bg-blue-100 rounded mt-2"
                >
                  + Add Item
                </button>
              </div>
            </div>

          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex w-full">
          <UiAlert v-if="formError" type="error">{{ formError }}</UiAlert>
        </div>
        <UiButton @click="showCreateModal = false" variant="secondary">Cancel</UiButton>
        <UiButton @click="handleCreate" :loading="submitting">Create</UiButton>
      </template>
    </UiModal>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import type { Document, CreateDocumentRequest } from '@/types';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiTable from '@/core/ui/components/UiTable.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiSelect from '@/core/ui/components/UiSelect.vue';
import UiComboBox from '@/core/ui/components/UiComboBox.vue';
import UiTextarea from '@/core/ui/components/UiTextarea.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';

const authStore = useAuthStore();
const isAdmin = computed(() => authStore.role === 'RoleAdmin');

const documents = ref<Document[]>([]);
const loading = ref(false);
const error = ref('');
const uuidSearchQuery = ref('');
const titleSearchQuery = ref('');
const offset = ref(0);
const limit = ref(15);

const filters = ref({
  folder_name: '',
  date_from: '',
  date_to: '',
  sort_by: '',
  field_key: '',
  field_value: ''
});

// ComboBox Search relationship
const fieldOptions = ['Nature', 'Status', 'Complainants', 'Respondents'];

const valueOptionsMap: any = {
  'Nature': ['Civil', 'Criminal'],
  'Status': ['Case Filed', 'Mediation', 'Arbitration', 'Conciliation', 'Pending', 'Resolved'],
};

// Computed property to handle the dynamic list for the second ComboBox Search
const dynamicValueOptions = computed(() => {
  const selectedKey = filters.value.field_key;
  return valueOptionsMap[selectedKey] || [];
});

// Reset the second box if the first one changes
watch(() => filters.value.field_key, () => {
  filters.value.field_value = '';
});

const sortOptions = [
  { value: 'created_at', label: 'Created Date' },
  { value: 'code', label: 'Code' },
  { value: 'title', label: 'Title' }
];

const showCreateModal = ref(false);
const submitting = ref(false);
const formError = ref('');

const form = ref<CreateDocumentRequest>({
  title: '',
  code: '',
  folder_name: '',
  fields: {
    nature: 'civil',
    status: 'case_filed', 
    complainants: [],
    respondents: [],
    complaint: '',
  }
});

const columns = [
  { key: 'code', label: 'Code' },
  { key: 'title', label: 'Title' },
  { key: 'status', label: 'Status'},
  { key: 'folder_name', label: 'Folder Name' },
  { key: 'files', label: 'Files' },
  { key: 'created_at', label: 'Created' },
  { key: 'actions', label: 'Actions' }
];

const parseSystemError = (err: any): string => {
  const rawError = err.response?.data?.error || '';
  
  // Specific check for physical folder mismatch
  if (rawError.includes('no such file or directory') || rawError.includes('scanning physical folder')) {
    return 'Physical directory mismatch detected. Please reload documents in the admin dashboard.';
  }
  
  return rawError || 'An unexpected system error occurred.';
};

async function loadDocuments() {
  loading.value = true;
  error.value = '';
  try {
    const response = await DocumentService.getAll(offset.value, limit.value);
    documents.value = response.data;
  } catch (err: any) {
    error.value = parseSystemError(err);
  } finally {
    loading.value = false;
  }
}

async function performSearch(resetOffset = true) {
  loading.value = true;
  error.value = '';

  // Only reset to page 1 if we are starting a brand new search
  if (resetOffset) {
    offset.value = 0;
  }
  
  try {
    const params: Record<string, any> = {
      code: uuidSearchQuery.value || undefined,
      title: titleSearchQuery.value || undefined,
      folder_name: filters.value.folder_name || undefined,
      date_from: filters.value.date_from || undefined,
      date_to: filters.value.date_to || undefined,
      sort_by: filters.value.sort_by || undefined,
      sort_desc: filters.value.sort_by ? true : undefined,
      offset: offset.value, // This will now correctly use the incremented value
      limit: limit.value
    };

    if (filters.value.field_key) {
      const key = filters.value.field_key.toLowerCase();
      const val = filters.value.field_value?.toLowerCase() ?? '';
      params[`field_${key}`] = val;
    }

    const response = await DocumentService.search(params);
    documents.value = response.data;
  } catch (err: any) {
    error.value = parseSystemError(err);
  } finally {
    loading.value = false;
  }
}

function resetSearch() {
  uuidSearchQuery.value = '';
  titleSearchQuery.value = '';
  filters.value = {
    folder_name: '',
    date_from: '',
    date_to: '',
    sort_by: '',
    field_key: '',
    field_value: ''
  };
  offset.value = 0;
  loadDocuments();
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString();
}

async function handleCreate() {
  formError.value = '';
  submitting.value = true;
  
  try {
    await DocumentService.create({
      title: form.value.title,
      fields: form.value.fields
    });
    showCreateModal.value = false;
    form.value = { title: '', code: '', folder_name: '', fields: {} };
    loadDocuments();
  } catch (err: any) {
    formError.value = err.response?.data?.error || 'Failed to create document';
  } finally {
    submitting.value = false;
  }
}

function nextPage() {
  offset.value += limit.value;
  isSearchActive() ? performSearch(false) : loadDocuments();
}

function prevPage() {
  offset.value = Math.max(0, offset.value - limit.value);
  isSearchActive() ? performSearch(false) : loadDocuments();
}

function isSearchActive(): boolean {
  return (
    uuidSearchQuery.value !== '' ||
    titleSearchQuery.value !== '' ||
    filters.value.folder_name !== '' ||
    filters.value.date_from !== '' ||
    filters.value.date_to !== '' ||
    filters.value.sort_by !== '' ||
    filters.value.field_key !== '' ||
    filters.value.field_value !== ''
  );
}

onMounted(() => {
  loadDocuments();
});
</script>

