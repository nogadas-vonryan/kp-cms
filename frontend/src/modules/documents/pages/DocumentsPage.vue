<template>
  <div class="space-y-4">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-900">Documents</h1>
      <UiButton v-if="isAdmin" @click="showCreateModal = true" class="w-full sm:w-auto flex items-center justify-center gap-2">
        <Plus :size="18" />
        <span>Create Document</span>
      </UiButton>
    </div>

    <!-- Search & Filters -->
    <UiCard>
      <form @submit.prevent="performSearch()" class="space-y-3">
        <!-- Search Input with Quick Actions -->
        <div class="flex gap-2">
          <div class="relative flex-1">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
            <input
              v-model="uuidSearchQuery"
              placeholder="Search by code..."
              class="w-full pl-9 pr-3 py-2.5 text-sm border border-gray-300 rounded-lg bg-white focus:ring focus:ring-blue-300 focus:border-transparent outline-none transition-all"
              @keyup.enter="performSearch()"
            />
          </div>
          <UiButton 
            type="submit" 
            variant="secondary"
            class="shrink-0 flex items-center justify-center gap-2"
          >
            <Search :size="16" />
            <span class="hidden sm:inline">Search</span>
          </UiButton>
        </div>

        <!-- Quick Sort & Filter Controls -->
        <div class="flex gap-2 flex-wrap items-center">
          <select
            v-model="filters.sort_by"
            class="flex-1 min-w-35 text-sm p-2 border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
          >
            <option value="">Sort by...</option>
            <option value="created_at">Created Date</option>
            <option value="code">Code</option>
            <option value="title">Title</option>
          </select>
          
          <button
            v-if="filters.sort_by"
            type="button"
            @click="toggleSortDirection"
            :title="filters.sort_desc ? 'Sort Descending' : 'Sort Ascending'"
            class="p-2.5 transition-colors flex items-center justify-center rounded-lg text-gray-800 bg-white hover:bg-gray-100 border border-gray-300"
          >
            <ArrowDown v-if="filters.sort_desc" :size="16" />
            <ArrowUp v-else :size="16" />
          </button>

          <UiButton 
            type="button" 
            @click="resetSearch"
            class="shrink-0 flex items-center justify-center gap-1"
          >
            <RotateCcw :size="16" />
            <span class="hidden sm:inline">Reset</span>
          </UiButton>
        </div>

        <!-- Advanced Filters Toggle -->
        <details class="group">
          <summary class="text-sm text-blue-600 hover:text-blue-700 cursor-pointer font-medium list-none flex items-center gap-2 select-none p-2 hover:bg-blue-50 rounded transition-colors">
            <ChevronRight :size="16" class="group-open:rotate-90 transition-transform" />
            <span>More Filters</span>
          </summary>
          
          <div class="pt-4 space-y-3 border-t border-gray-100 mt-2">
            <!-- Title Search -->
            <div>
              <label class="text-xs font-semibold text-gray-600 uppercase tracking-wider mb-1.5 block">Title</label>
              <input
                v-model="titleSearchQuery"
                placeholder="Search by title..."
                class="w-full text-sm p-2 border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
              />
            </div>

            <!-- Folder Filter -->
            <div>
              <label class="text-xs font-semibold text-gray-600 uppercase tracking-wider mb-1.5 block">Folder</label>
              <input
                v-model="filters.folder_name"
                placeholder="Filter by folder..."
                class="w-full text-sm p-2 border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
              />
            </div>

            <!-- Date Range -->
            <div>
              <label class="text-xs font-semibold text-gray-600 uppercase tracking-wider mb-1.5 block">Date Range</label>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <input
                  v-model="filters.date_from"
                  type="date"
                  @change="performSearch()"
                  class="text-sm p-2 border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
                />
                <input
                  v-model="filters.date_to"
                  type="date"
                  @change="performSearch()"
                  class="text-sm p-2 border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
                />
              </div>
            </div>

            <!-- Field Filters -->
            <div>
              <label class="text-xs font-semibold text-gray-600 uppercase tracking-wider mb-1.5 block">Document Field</label>
              <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                <select
                  v-model="filters.field_key"
                  class="text-sm p-2 border border-gray-300 rounded-lg bg-white focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
                >
                  <option value="">Select field...</option>
                  <option v-for="opt in fieldOptions" :key="opt" :value="opt">{{ opt }}</option>
                </select>

                <UiComboBox
                  v-model="filters.field_value"
                  :options="dynamicValueOptions"
                  :placeholder="filters.field_key ? `Enter ${filters.field_key}...` : 'Select value...'"
                  @option-selected="handleFieldValueSelected"
                />
              </div>
            </div>

            <!-- Active Filters Summary -->
            <div v-if="hasActiveFilters" class="p-2 bg-blue-50 rounded-lg">
              <p class="text-xs text-blue-700 font-medium">
                {{ getActiveFiltersCount }} filter{{ getActiveFiltersCount !== 1 ? 's' : '' }} applied
              </p>
            </div>
          </div>
        </details>
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
      
      <div class="w-full overflow-x-auto">
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

        <!-- Mobile Card View -->
        <div v-if="documents.length > 0 && isMobileView" class="divide-y divide-gray-200 border-t border-gray-200">
          <div 
            v-for="(doc) in documents" 
            :key="doc.uuid"
            @click="router.push(`/documents/${doc.uuid}`)"
            class="p-4 hover:bg-gray-50 transition-colors cursor-pointer space-y-2"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <span class="font-mono text-xs text-gray-500">{{ doc.code }}</span>
                <p class="font-medium text-gray-900 wrap-break-word line-clamp-2">{{ doc.title }}</p>
              </div>
              <span 
                :class="getStatusBadgeClass(doc.fields?.status || 'none')"
                class="text-[10px] font-medium px-2 py-1 rounded-full shrink-0 whitespace-nowrap"
              >
                {{ formatStatusLabel(doc.fields?.status || 'none') }}
              </span>
            </div>
            <div class="grid grid-cols-2 gap-2 text-xs text-gray-600">
              <div>
                <span class="text-gray-500">Folder:</span> {{ doc.folder_name }}
              </div>
              <div>
                <span class="text-gray-500">Files:</span> {{ doc.files.length }}
              </div>
              <div class="col-span-2">
                <span class="text-gray-500">Created:</span> {{ formatDate(doc.created_at) }}
              </div>
            </div>
          </div>
        </div>

        <!-- Desktop Table View -->
        <div v-else-if="documents.length > 0" class="table-wrapper" @click="handleRowClick">
          <table class="w-full text-sm">
            <thead class="border-b border-gray-200 bg-gray-50">
              <tr>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Code</th>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Title</th>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Status</th>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Folder</th>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Files</th>
                <th class="text-left px-4 py-3 font-semibold text-gray-900">Created</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr v-for="doc in documents" :key="doc.uuid" class="hover:bg-gray-50 transition-colors cursor-pointer">
                <td class="px-4 py-3"><span class="font-mono text-xs text-gray-600">{{ doc.code }}</span></td>
                <td class="px-4 py-3"><span class="text-gray-600 truncate block max-w-xs">{{ doc.title }}</span></td>
                <td class="px-4 py-3">
                  <span 
                    :class="getStatusBadgeClass(doc.fields?.status || 'none')"
                    class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium"
                  >
                    {{ formatStatusLabel(doc.fields?.status || 'none') }}
                  </span>
                </td>
                <td class="px-4 py-3"><span class="text-gray-600 truncate block max-w-xs">{{ doc.folder_name }}</span></td>
                <td class="px-4 py-3"><span class="text-gray-600">{{ doc.files.length }}</span></td>
                <td class="px-4 py-3"><span class="text-gray-600">{{ formatDate(doc.created_at) }}</span></td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-else-if="!loading" class="p-8 text-center text-gray-600">
          No documents found
        </div>
      </div>
    </UiCard>

    <!-- Pagination -->
    <div v-if="documents.length > 0" class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
      <span class="text-xs sm:text-sm text-gray-600">
        Showing {{ offset + 1 }}-{{ Math.min(offset + limit, offset + documents.length) }}
      </span>
      <div class="flex gap-2 w-full sm:w-auto">
        <UiButton 
          @click="prevPage" 
          :disabled="offset === 0"
          class="flex-1 sm:flex-none flex items-center justify-center gap-1"
        >
          <ChevronLeft :size="16" />
          <span class="hidden sm:inline">Previous</span>
        </UiButton>
        <UiButton 
          @click="nextPage" 
          :disabled="documents.length < limit"
          class="flex-1 sm:flex-none flex items-center justify-center gap-1"
        >
          <span class="hidden sm:inline">Next</span>
          <ChevronRight :size="16" />
        </UiButton>
      </div>
    </div>

    <!-- Create Modal -->
    <UiModal v-model:open="showCreateModal" title="Create Document">
      <form @submit.prevent="handleCreate" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Title <span class="text-red-500">*</span></label>
          <input v-model="form.title" required class="w-full text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Code <span v-if="isCodeRequired" class="text-red-500">*</span>
          </label>
          <input
            v-model="form.code"
            :required="isCodeRequired"
            :placeholder="isCodeRequired ? 'Document code (required)' : 'Document code (optional)'"
            class="w-full text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Created At</label>
          <input v-model="form.created_at" type="date" class="w-full text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>

        <div class="pt-4 border-t border-gray-200">
          <h3 class="font-semibold text-gray-900 mb-3 text-sm">Document Details</h3>
          
          <div class="space-y-4">
            
            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Nature</span>
              </div>
              <select 
                v-model="form.fields.nature"
                class="w-full text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none"
              >
                <option value="civil">Civil</option>
                <option value="criminal">Criminal</option>
              </select>
            </div>

            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Complaint</span>
              </div>
              <textarea v-model="form.fields.complaint" rows="3" class="w-full text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none"></textarea>
            </div>

            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Complainants</span>
              </div>
              <div class="space-y-2">
                <div v-for="(_, index) in form.fields.complainants" :key="index" class="flex gap-2 items-start">
                  <input v-model="form.fields.complainants![index]" class="flex-1 text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none" placeholder="Name" />
                  <button 
                    type="button" 
                    @click="form.fields.complainants?.splice(index, 1)"
                    class="p-2 text-red-600 hover:bg-red-600 hover:text-white rounded transition-colors shrink-0"
                    title="Remove"
                  >
                    <Trash2 :size="16" />
                  </button>
                </div>
                <UiButton
                  type="button"
                  @click="form.fields.complainants?.push('')"
                  block
                  class="flex items-center justify-center gap-1 text-xs"
                >
                  <Plus :size="14" />
                  <span>Add Complainant</span>
                </UiButton>
              </div>
            </div>

            <div class="rounded p-3 border border-gray-200">
              <div class="flex items-center justify-between mb-2">
                <span class="font-medium text-gray-900 text-sm">Respondents</span>
              </div>
              <div class="space-y-2">
                <div v-for="(_, index) in form.fields.respondents" :key="index" class="flex gap-2 items-start">
                  <input v-model="form.fields.respondents![index]" class="flex-1 text-sm p-2 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none" placeholder="Name" />
                  <button 
                    type="button" 
                    @click="form.fields.respondents?.splice(index, 1)"
                    class="p-2 text-red-600 hover:bg-red-600 hover:text-white rounded transition-colors shrink-0"
                    title="Remove"
                  >
                    <Trash2 :size="16" />
                  </button>
                </div>
                <UiButton
                  type="button"
                  @click="form.fields.respondents?.push('')"
                  block
                  class="flex items-center justify-center gap-1 text-xs"
                >
                  <Plus :size="14" />
                  <span>Add Respondent</span>
                </UiButton>
              </div>
            </div>

          </div>
        </div>
      </form>

      <template #footer>
        <div class="flex w-full flex-col gap-2">
          <UiAlert v-if="formError" type="error" class="max-h-24 overflow-y-auto">
            <span class="block text-sm wrap-break-word">{{ formError }}</span>
          </UiAlert>
          <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <UiButton @click="showCreateModal = false">Cancel</UiButton>
            <UiButton @click="handleCreate" :disabled="submitting" variant="secondary" class="flex items-center gap-2">
              <Check :size="16" />
              <span>{{ submitting ? 'Creating...' : 'Create' }}</span>
            </UiButton>
          </div>
        </div>
      </template>
    </UiModal>

  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue';
import { useRouter } from 'vue-router';
import { Plus, Search, RotateCcw, ArrowUp, ArrowDown, ChevronLeft, ChevronRight, Trash2, Check } from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { DocumentService } from '@/modules/documents/services/documentService';
import { extractErrorMessage } from '@/core/api';
import type { Document, CreateDocumentRequest } from '@/types';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiComboBox from '@/core/ui/components/UiComboBox.vue';

const authStore = useAuthStore();
const router = useRouter();
const isAdmin = computed(() => authStore.role === 'RoleAdmin');

const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1024);
const isMobileView = computed(() => windowWidth.value < 1024);

onMounted(() => {
  const handleResize = () => {
    windowWidth.value = window.innerWidth;
  };
  window.addEventListener('resize', handleResize);
  return () => window.removeEventListener('resize', handleResize);
});

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
  sort_desc: true,
  field_key: '',
  field_value: ''
});

const fieldOptions = ['Nature', 'Status', 'Complainants', 'Respondents'];

// Display values (user-facing)
const valueOptionsMap: any = {
  'Nature': ['Civil', 'Criminal'],
  'Status': ['Case Filed', 'Mediation', 'Arbitration', 'Conciliation', 'Repudiation', 'Pending', 'Resolved'],
};

// Mapping for API conversions to snake_case
const apiValueMap: Record<string, Record<string, string>> = {
  'Nature': {
    'Civil': 'civil',
    'Criminal': 'criminal'
  },
  'Status': {
    'Case Filed': 'case_filed',
    'Mediation': 'mediation',
    'Arbitration': 'arbitration',
    'Conciliation': 'conciliation',
    'Repudiation': 'repudiation',
    'Pending': 'pending',
    'Resolved': 'resolved'
  }
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

// Compute active filters for display
const hasActiveFilters = computed(() => {
  return (
    uuidSearchQuery.value !== '' ||
    titleSearchQuery.value !== '' ||
    filters.value.folder_name !== '' ||
    filters.value.date_from !== '' ||
    filters.value.date_to !== '' ||
    filters.value.sort_by !== '' ||
    filters.value.field_key !== '' ||
    filters.value.field_value !== '' ||
    filters.value.sort_by !== ''
  );
});

const getActiveFiltersCount = computed(() => {
  let count = 0;
  if (uuidSearchQuery.value !== '') count++;
  if (titleSearchQuery.value !== '') count++;
  if (filters.value.folder_name !== '') count++;
  if (filters.value.date_from !== '') count++;
  if (filters.value.date_to !== '') count++;
  if (filters.value.field_key !== '') count++;
  if (filters.value.field_value !== '') count++;
  if (filters.value.sort_by !== '') count++;
  return count;
});

const showCreateModal = ref(false);
const submitting = ref(false);
const formError = ref('');

const isCodeRequired = computed(() => Boolean(form.value.created_at));

const form = ref<CreateDocumentRequest>({
  title: '',
  code: '',
  created_at: '',
  folder_name: '',
  fields: {
    nature: 'civil',
    status: 'case_filed', 
    complainants: [],
    respondents: [],
    complaint: '',
  }
});

const parseSystemError = (err: any): string => {
  const extracted = extractErrorMessage(err);
  const rawError = err?.response?.data?.error || '';

  // Specific check for physical folder mismatch (prefer this specific message)
  if (rawError.includes('no such file or directory') || rawError.includes('scanning physical folder')) {
    return 'Physical directory mismatch detected. Please reload documents in the admin dashboard.';
  }

  return extracted || rawError || 'An unexpected system error occurred.';
};

async function loadDocuments() {
  loading.value = true;
  error.value = '';
  try {
    const response = await DocumentService.getAll(offset.value, limit.value, 'created_at', true);
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
      sort_desc: filters.value.sort_by ? filters.value.sort_desc : undefined,
      offset: offset.value, // This will now correctly use the incremented value
      limit: limit.value
    };

    if (filters.value.field_key) {
      // Convert field key to snake_case
      const key = filters.value.field_key
        .replace(/\s+/g, '_')
        .replace(/([A-Z])/g, '_$1')
        .toLowerCase()
        .replace(/^_/, '');
      // Convert display value to API snake_case value
      const apiValue = filters.value.field_value 
        ? (apiValueMap[filters.value.field_key]?.[filters.value.field_value] ?? filters.value.field_value.toLowerCase())
        : '';
      params[`field_${key}`] = apiValue;
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
    sort_desc: true,
    field_key: '',
    field_value: ''
  };
  offset.value = 0;
  loadDocuments();
}

function formatDate(dateStr: string) {
  return new Date(dateStr).toLocaleDateString();
}

function getStatusBadgeClass(status: string): string {
  const statusClasses: Record<string, string> = {
    'case_filed': 'bg-blue-100 text-blue-800',
    'mediation': 'bg-purple-100 text-purple-800',
    'arbitration': 'bg-orange-100 text-orange-800',
    'conciliation': 'bg-teal-100 text-teal-800',
    'repudiation': 'bg-red-100 text-red-800',
    'resolved': 'bg-green-100 text-green-800',
    'pending': 'bg-yellow-100 text-yellow-800',
    'none': 'bg-gray-100 text-gray-800'
  };
  return statusClasses[status] || 'bg-gray-100 text-gray-800';
}

function formatStatusLabel(status: string): string {
  const labels: Record<string, string> = {
    'case_filed': 'Case Filed',
    'mediation': 'Mediation',
    'arbitration': 'Arbitration',
    'conciliation': 'Conciliation',
    'repudiation': 'Repudiation',
    'resolved': 'Resolved',
    'pending': 'Pending',
    'none': 'None'
  };
  return labels[status] || status;
}

async function handleCreate() {
  formError.value = '';
  submitting.value = true;

  if (form.value.created_at && !form.value.code?.trim()) {
    formError.value = 'Code is required when Created At is set.';
    submitting.value = false;
    return;
  }
  
  try {
    const response = await DocumentService.create({
      title: form.value.title,
      code: form.value.code,
      created_at: form.value.created_at ? new Date(form.value.created_at).toISOString() : '',
      fields: form.value.fields
    });
    showCreateModal.value = false;
    form.value = { title: '', code: '', created_at: '', folder_name: '', fields: {
      nature: 'civil', status: 'case_filed', complainants: [], respondents: [],
    } };
    // Redirect to the newly created document
    if (response.data?.uuid) {
      router.push(`/documents/${response.data.uuid}`);
    } else {
      loadDocuments();
    }
  } catch (err: any) {
    formError.value = extractErrorMessage(err) || 'Failed to create document';
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

function toggleSortDirection() {
  filters.value.sort_desc = !filters.value.sort_desc;
  performSearch();
}

function handleFieldValueSelected() {
  // Use nextTick to ensure the v-model is synced before performing search
  nextTick(() => performSearch());
}

function handleRowClick(event: MouseEvent) {
  const target = event.target as HTMLElement;
  if (target.closest('[role="button"]') || target.closest('a')) {
    return;
  }
  
  const table = (event.currentTarget as HTMLElement).closest('.w-full');
  if (!table) return;
  
  const rows = Array.from(table.querySelectorAll('tbody tr'));
  const clickedRow = rows.find(row => row.contains(target));
  
  if (clickedRow) {
    const doc = documents.value[rows.indexOf(clickedRow)];
    if (doc) {
      router.push(`/documents/${doc.uuid}`);
    }
  }
}

onMounted(() => {
  loadDocuments();
});
</script>

<style scoped>
.table-wrapper {
  cursor: pointer;
}

.table-wrapper :deep(tbody tr) {
  transition: background-color 0.15s ease;
}

.table-wrapper :deep(tbody tr:hover) {
  background-color: #f3f4f6;
}
</style>

