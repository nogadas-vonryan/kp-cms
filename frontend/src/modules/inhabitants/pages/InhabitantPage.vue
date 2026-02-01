<template>
  <div class="space-y-4">
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-900">Inhabitants</h1>
      <UiButton v-if="isAdmin" @click="openCreateModal" class="w-full sm:w-auto flex items-center justify-center gap-2">
        <Plus :size="18" />
        <span>Add Inhabitant</span>
      </UiButton>
    </div>

    <UiCard>
      <div class="space-y-3">
        <div class="flex gap-2">
          <div class="relative flex-1">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" :size="16" />
            <input
              v-model="searchQuery"
              placeholder="Search by name..."
              class="w-full pl-9 pr-3 py-2.5 text-sm border border-gray-300 rounded-lg bg-white focus:ring focus:ring-blue-300 focus:border-transparent outline-none transition-all"
              @keyup.enter="handleSearch"
            />
          </div>
          <UiButton 
            @click="handleSearch"
            variant="secondary"
            class="shrink-0 flex items-center justify-center gap-2"
          >
            <Search :size="16" />
            <span class="hidden sm:inline">Search</span>
          </UiButton>
          <UiButton 
            @click="resetSearch"
            class="shrink-0 flex items-center justify-center gap-1"
          >
            <RotateCcw :size="16" />
            <span class="hidden sm:inline">Reset</span>
          </UiButton>
        </div>
      </div>
    </UiCard>

    <UiCard :padded="false" class="relative min-h-125">
      <div 
        v-if="loading" 
        class="absolute inset-0 z-20 flex items-center justify-center bg-white/60 backdrop-blur-[1px]"
      >
        <div class="flex flex-col items-center gap-2">
          <span class="text-sm font-medium text-gray-600">Loading inhabitants...</span>
        </div>
      </div>
      
      <div class="w-full overflow-x-auto">
        <UiSystemNotice v-if="error" type="error" label="System_Error" :modelValue="true" class="m-4">
          <template #title>
            <span class="px-2 text-sm text-gray-700 leading-relaxed">{{ error }}</span>
          </template>
        </UiSystemNotice>

        <div v-if="filteredInhabitants.length > 0 && isMobileView" class="divide-y divide-gray-200 border-t border-gray-200">
          <div 
            v-for="inh in filteredInhabitants" 
            :key="inh.id"
            class="p-4 hover:bg-gray-50 transition-colors cursor-pointer space-y-2"
            @click="viewDetails(inh)"
          >
            <div class="flex items-start justify-between gap-2">
              <div class="flex-1 min-w-0">
                <p class="font-medium text-gray-900">
                  {{ inh.last_name }}, {{ inh.first_name }} {{ inh.suffix }}
                </p>
                <p class="text-xs text-gray-500">{{ inh.contact_no || 'No contact info' }}</p>
              </div>
              <ChevronRight :size="16" class="text-gray-400" />
            </div>
            <div class="text-xs text-gray-600 truncate">
              <span class="text-gray-500">Address:</span> {{ inh.address }}
            </div>
          </div>
        </div>

        <div v-else-if="filteredInhabitants.length > 0" class="table-wrapper">
          <table class="w-full text-sm text-left">
            <thead class="border-b border-gray-200 bg-gray-50">
              <tr>
                <th class="px-4 py-3 font-semibold text-gray-900">Name</th>
                <th class="px-4 py-3 font-semibold text-gray-900">Birthday</th>
                <th class="px-4 py-3 font-semibold text-gray-900">Contact</th>
                <th class="px-4 py-3 font-semibold text-gray-900">Address</th>
                <th v-if="isAdmin" class="px-4 py-3 font-semibold text-gray-900 text-right">Actions</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr 
                v-for="inh in filteredInhabitants" 
                :key="inh.id" 
                class="hover:bg-gray-50 transition-colors"
              >
                <td class="px-4 py-3 font-medium text-gray-900">
                  {{ inh.last_name }}, {{ inh.first_name }} {{ inh.middle_name }} {{ inh.suffix }}
                </td>
                <td class="px-4 py-3 text-gray-600">{{ inh.birthday }}</td>
                <td class="px-4 py-3 text-gray-600">{{ inh.contact_no }}</td>
                <td class="px-4 py-3 text-gray-600 truncate max-w-xs">{{ inh.address }}</td>
                <td v-if="isAdmin" class="px-4 py-3 text-right">
                  <div class="flex justify-end gap-2">
                    <button @click.stop="editInhabitant(inh)" class="p-1.5 hover:bg-blue-50 text-blue-600 rounded">
                      <Edit3 :size="16" />
                    </button>
                    <button @click.stop="confirmDelete(inh)" class="p-1.5 hover:bg-red-50 text-red-600 rounded">
                      <Trash2 :size="16" />
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-else-if="!loading" class="p-8 text-center text-gray-600">
          No inhabitants found
        </div>
      </div>
    </UiCard>

    <div v-if="inhabitants.length > 0" class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
      <span class="text-xs sm:text-sm text-gray-600">
        Showing {{ offset + 1 }}-{{ Math.min(offset + limit, inhabitants.length) }}
      </span>
      <div class="flex gap-2 w-full sm:w-auto">
        <UiButton @click="prevPage" :disabled="offset === 0" class="flex-1 sm:flex-none flex items-center justify-center gap-1">
          <ChevronLeft :size="16" />
          <span>Previous</span>
        </UiButton>
        <UiButton @click="nextPage" :disabled="inhabitants.length < limit" class="flex-1 sm:flex-none flex items-center justify-center gap-1">
          <span>Next</span>
          <ChevronRight :size="16" />
        </UiButton>
      </div>
    </div>

    <UiModal v-model:open="showModal" :title="isEditing ? 'Edit Inhabitant' : 'Add New Inhabitant'">
      <form @submit.prevent="handleSubmit" class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">First Name <span class="text-red-500">*</span></label>
          <input v-model="form.first_name" required class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Last Name <span class="text-red-500">*</span></label>
          <input v-model="form.last_name" required class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Middle Name</label>
          <input v-model="form.middle_name" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Suffix</label>
          <input v-model="form.suffix" placeholder="e.g. Jr., III" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Birthday</label>
          <input v-model="form.birthday" type="date" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Contact No.</label>
          <input v-model="form.contact_no" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-2">
          <label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
          <textarea v-model="form.address" rows="2" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"></textarea>
        </div>
      </form>

      <template #footer>
        <div class="flex w-full flex-col gap-2">
          <UiAlert v-if="formError" type="error">{{ formError }}</UiAlert>
          <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <UiButton @click="showModal = false">Cancel</UiButton>
            <UiButton @click="handleSubmit" :disabled="submitting" variant="secondary" class="flex items-center gap-2">
              <Check :size="16" />
              <span>{{ submitting ? 'Saving...' : 'Save' }}</span>
            </UiButton>
          </div>
        </div>
      </template>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { 
  Plus, Search, RotateCcw, ChevronLeft, ChevronRight, 
  Trash2, Edit3, Check 
} from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { InhabitantService, type Inhabitant } from '@/modules/inhabitants/services/inhabitantService';
import { extractErrorMessage } from '@/core/api';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';
import UiButton from '@/core/ui/components/UiButton.vue';

const authStore = useAuthStore();
const isAdmin = computed(() => authStore.role === 'RoleAdmin');

// Layout Responsiveness
const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 1024);
const isMobileView = computed(() => windowWidth.value < 1024);

onMounted(() => {
  const handleResize = () => { windowWidth.value = window.innerWidth; };
  window.addEventListener('resize', handleResize);
  loadInhabitants();
});

// State
const inhabitants = ref<Inhabitant[]>([]);
const loading = ref(false);
const error = ref('');
const searchQuery = ref('');
const offset = ref(0);
const limit = ref(15);

// Modal & Form State
const showModal = ref(false);
const isEditing = ref(false);
const submitting = ref(false);
const formError = ref('');
const currentInhabitantId = ref<number | null>(null);

const form = ref<Inhabitant>({
  first_name: '',
  last_name: '',
  middle_name: '',
  suffix: '',
  birthday: '',
  contact_no: '',
  address: ''
});

// Computed Search (Client-side search for current page, or could be integrated into API)
const filteredInhabitants = computed(() => {
  if (!searchQuery.value) return inhabitants.value;
  const q = searchQuery.value.toLowerCase();
  return inhabitants.value.filter(i => 
    i.first_name.toLowerCase().includes(q) || 
    i.last_name.toLowerCase().includes(q)
  );
});

async function loadInhabitants() {
  loading.value = true;
  error.value = '';
  try {
    const response = await InhabitantService.getAll(offset.value, limit.value);
    inhabitants.value = response.data;
  } catch (err: any) {
    error.value = extractErrorMessage(err) || 'Failed to load inhabitants';
  } finally {
    loading.value = false;
  }
}

function handleSearch() {
  offset.value = 0;
  loadInhabitants();
}

function resetSearch() {
  searchQuery.value = '';
  offset.value = 0;
  loadInhabitants();
}

// Actions
function openCreateModal() {
  isEditing.value = false;
  currentInhabitantId.value = null;
  form.value = { first_name: '', last_name: '', middle_name: '', suffix: '', birthday: '', contact_no: '', address: '' };
  formError.value = '';
  showModal.value = true;
}

function editInhabitant(inh: Inhabitant) {
  isEditing.value = true;
  currentInhabitantId.value = inh.id!;
  form.value = { ...inh };
  formError.value = '';
  showModal.value = true;
}

async function handleSubmit() {
  submitting.value = true;
  formError.value = '';
  try {
    if (isEditing.value && currentInhabitantId.value) {
      await InhabitantService.update(currentInhabitantId.value, form.value);
    } else {
      await InhabitantService.create(form.value);
    }
    showModal.value = false;
    loadInhabitants();
  } catch (err: any) {
    formError.value = extractErrorMessage(err) || 'Failed to save inhabitant';
  } finally {
    submitting.value = false;
  }
}

async function confirmDelete(inh: Inhabitant) {
  if (!confirm(`Are you sure you want to delete ${inh.first_name} ${inh.last_name}?`)) return;
  try {
    await InhabitantService.delete(inh.id!);
    loadInhabitants();
  } catch (err: any) {
    alert(extractErrorMessage(err) || 'Failed to delete inhabitant');
  }
}

function nextPage() {
  offset.value += limit.value;
  loadInhabitants();
}

function prevPage() {
  offset.value = Math.max(0, offset.value - limit.value);
  loadInhabitants();
}

function viewDetails(inh: Inhabitant) {
  if (isAdmin.value) editInhabitant(inh);
}
</script>

<style scoped>
.table-wrapper :deep(tbody tr) {
  transition: background-color 0.15s ease;
}
</style>