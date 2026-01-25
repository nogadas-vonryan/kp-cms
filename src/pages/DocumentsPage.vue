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
      <form @submit.prevent="performSearch">
        <div class="space-y-3">
          <div class="flex gap-2">
            <UiInput
              v-model="searchQuery"
              placeholder="Search by code..."
              class="flex-1"
            />
            <UiInput
              v-model="filters.folder_name"
              placeholder="Filter by folder..."
              class="flex-1"
            />
            <UiButton type="submit">Search</UiButton>
            <UiButton type="button" @click="resetSearch" variant="secondary">Reset</UiButton>
          </div>
          
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
            <UiInput
              v-model="filters.field_key"
              placeholder="Field name (e.g. status)"
            />
            <UiInput
              v-model="filters.field_value"
              placeholder="Field value (optional)"
            />
          </div>
          <div class="grid grid-cols-1 gap-3">
            <UiSelect
              v-model="filters.sort_by"
              :options="sortOptions"
              placeholder="Sort by..."
            />
          </div>
        </div>
      </form>
    </UiCard>

    <!-- Documents Table -->
    <UiCard :padded="false">
      <div v-if="loading" class="p-8 text-center text-gray-600">
        Loading documents...
      </div>
      
      <UiAlert v-else-if="error" type="error" class="m-4">
        {{ error }}
      </UiAlert>

      <UiTable
        v-else-if="documents.length > 0"
        :columns="columns"
        :rows="documents"
      >
        <template #cell:code="{ value }">
          <span class="font-mono text-sm">{{ value }}</span>
        </template>

        <template #cell:folder_name="{ value }">
          <span class="text-sm text-gray-600">{{ value }}</span>
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

      <div v-else class="p-8 text-center text-gray-600">
        No documents found
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
          <label class="block text-sm font-medium text-gray-700 mb-1">Title *</label>
          <UiInput v-model="form.title" required />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Code</label>
          <UiInput v-model="form.code" placeholder="Auto-generated if empty" />
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Folder Name</label>
          <UiInput v-model="form.folder_name" />
        </div>
        <UiAlert v-if="formError" type="error">{{ formError }}</UiAlert>
      </form>
      <template #footer>
        <UiButton @click="showCreateModal = false">Cancel</UiButton>
        <UiButton @click="handleCreate" :loading="submitting">Create</UiButton>
      </template>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
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

const authStore = useAuthStore();
const isAdmin = computed(() => authStore.role === 'RoleAdmin');

const documents = ref<Document[]>([]);
const loading = ref(false);
const error = ref('');
const searchQuery = ref('');
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
  fields: {}
});

const columns = [
  { key: 'code', label: 'Code' },
  { key: 'title', label: 'Title' },
  { key: 'folder_name', label: 'Folder Name' },
  { key: 'files', label: 'Files' },
  { key: 'created_at', label: 'Created' },
  { key: 'actions', label: 'Actions' }
];

async function loadDocuments() {
  loading.value = true;
  error.value = '';
  try {
    const response = await DocumentService.getAll(offset.value, limit.value);
    documents.value = response.data;
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to load documents';
  } finally {
    loading.value = false;
  }
}

async function performSearch() {
  loading.value = true;
  error.value = '';
  offset.value = 0;
  
  try {
    const params: Record<string, any> = {
      code: searchQuery.value || undefined,
      folder_name: filters.value.folder_name || undefined,
      date_from: filters.value.date_from || undefined,
      date_to: filters.value.date_to || undefined,
      sort_by: filters.value.sort_by || undefined,
      sort_desc: filters.value.sort_by ? true : undefined,
      offset: offset.value,
      limit: limit.value
    };

    if (filters.value.field_key) {
      params[`field_${filters.value.field_key}`] = filters.value.field_value ?? '';
    }

    const response = await DocumentService.search(params);
    documents.value = response.data;
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to search documents';
  } finally {
    loading.value = false;
  }
}

function resetSearch() {
  searchQuery.value = '';
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
      code: form.value.code || null,
      folder_name: form.value.folder_name || null,
      fields: {}
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
  isSearchActive() ? performSearch() : loadDocuments();
}

function prevPage() {
  offset.value = Math.max(0, offset.value - limit.value);
  isSearchActive() ? performSearch() : loadDocuments();
}

function isSearchActive(): boolean {
  return (
    searchQuery.value !== '' ||
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

