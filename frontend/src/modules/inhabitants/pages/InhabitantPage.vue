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
                <p class="text-xs text-gray-500">{{ formatField(inh.contact_no) }}</p>
              </div>
              <ChevronRight :size="16" class="text-gray-400" />
            </div>
            <div class="text-xs text-gray-600 truncate">
              <span class="text-gray-500">Address:</span> {{ formatField(inh.address) }}
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
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr 
                v-for="inh in filteredInhabitants" 
                :key="inh.id" 
                class="hover:bg-gray-50 transition-colors cursor-pointer"
                @click="viewDetails(inh)"
              >
                <td class="px-4 py-3 font-medium text-gray-900">
                  {{ inh.last_name }}, {{ inh.first_name }} {{ inh.middle_name }} {{ inh.suffix }}
                </td>
                <td class="px-4 py-3 text-gray-600">{{ formatBirthday(inh.birthday) }}</td>
                <td class="px-4 py-3 text-gray-600">{{ formatField(inh.contact_no) }}</td>
                <td class="px-4 py-3 text-gray-600 truncate max-w-xs">{{ formatField(inh.address) }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <div v-else-if="!loading" class="p-8 text-center text-gray-600">
          No inhabitants found
        </div>
      </div>
    </UiCard>

    <div v-if="inhabitants.length > 0 || offset > 0" class="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
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

    <!-- Create Inhabitant Modal -->
    <UiModal v-model:open="showCreateModal" title="Add New Inhabitant">
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
          <input v-model="form.birthday" type="date" :min="minBirthday" :max="maxBirthday" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
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
            <UiButton @click="showCreateModal = false">Cancel</UiButton>
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
import { useRouter } from 'vue-router';
import { 
  Plus, Search, RotateCcw, ChevronLeft, ChevronRight, Check
} from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { InhabitantService, type Inhabitant } from '@/modules/inhabitants/services/inhabitantService';
import { extractErrorMessage } from '@/core/api';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';
import UiButton from '@/core/ui/components/UiButton.vue';

const router = useRouter();
const authStore = useAuthStore();
const isAdmin = computed(() => authStore.role === 'RoleAdmin');

// Birthday constraints
const minBirthday = computed(() => {
  const date = new Date();
  date.setFullYear(date.getFullYear() - 120);
  return date.toISOString().split('T')[0];
});

const maxBirthday = computed(() => {
  const date = new Date();
  return date.toISOString().split('T')[0];
});

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
const showCreateModal = ref(false);
const submitting = ref(false);
const formError = ref('');
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

function openCreateModal() {
  form.value = {
    first_name: '',
    last_name: '',
    middle_name: '',
    suffix: '',
    birthday: '',
    contact_no: '',
    address: ''
  };
  formError.value = '';
  showCreateModal.value = true;
}

function validateBirthday(birthday: string): string | null {
  if (!birthday || birthday.trim() === '') {
    return null; // Empty is valid (optional field)
  }

  const birthdayDate = new Date(birthday);
  const minDate = new Date();
  minDate.setFullYear(minDate.getFullYear() - 130);
  const maxDate = new Date();

  if (birthdayDate < minDate) {
    return 'Date out of range';
  }

  if (birthdayDate > maxDate) {
    return 'Birthday cannot be in the future';
  }

  // Check for invalid dates like 0001-01-01
  if (birthday.startsWith('0000') || birthday.startsWith('0001')) {
    return 'Please enter a valid birthday';
  }

  return null; // Valid
}

async function handleSubmit() {
  // Validate birthday
  const birthdayError = validateBirthday(form.value.birthday);
  if (birthdayError) {
    formError.value = birthdayError;
    return;
  }

  submitting.value = true;
  formError.value = '';
  try {
    await InhabitantService.create(form.value);
    showCreateModal.value = false;
    loadInhabitants();
  } catch (err: any) {
    formError.value = extractErrorMessage(err) || 'Failed to create inhabitant';
  } finally {
    submitting.value = false;
  }
}

function formatField(value: string | undefined | null): string {
  if (!value || value.trim() === '') {
    return 'N/A';
  }
  return value;
}

function formatBirthday(birthday: string | undefined | null): string {
  if (!birthday || birthday.trim() === '') {
    return 'N/A';
  }
  
  // Check for invalid dates like 0001-01-01, 0000-00-00, etc.
  const invalidDates = ['0001-01-01', '0000-00-00', '1900-01-01'];
  if (invalidDates.includes(birthday)) {
    return 'N/A';
  }
  
  // Check if date starts with 0000 or 0001
  if (birthday.startsWith('0000') || birthday.startsWith('0001')) {
    return 'N/A';
  }
  
  return birthday;
}

function viewDetails(inh: Inhabitant) {
  router.push({ name: 'inhabitant-detail', params: { id: inh.id!.toString() } });
}

function nextPage() {
  offset.value += limit.value;
  loadInhabitants();
}

function prevPage() {
  offset.value = Math.max(0, offset.value - limit.value);
  loadInhabitants();
}
</script>

<style scoped>
.table-wrapper :deep(tbody tr) {
  transition: background-color 0.15s ease;
}
</style>