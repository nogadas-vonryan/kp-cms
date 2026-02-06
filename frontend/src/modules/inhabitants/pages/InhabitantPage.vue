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
            :key="inh.uuid"
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
                <th class="px-4 py-3 font-semibold text-gray-900">Birthdate</th>
                <th class="px-4 py-3 font-semibold text-gray-900">Contact</th>
                <th class="px-4 py-3 font-semibold text-gray-900">Address</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-gray-200">
              <tr 
                v-for="inh in filteredInhabitants" 
                :key="inh.uuid" 
                class="hover:bg-gray-50 transition-colors cursor-pointer"
                @click="viewDetails(inh)"
              >
                <td class="px-4 py-3 font-medium text-gray-900">
                  {{ inh.last_name }}, {{ inh.first_name }} {{ inh.middle_name }} {{ inh.suffix }}
                </td>
                <td class="px-4 py-3 text-gray-600">{{ formatBirthdate(inh.birthdate) }}</td>
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
          <label class="block text-sm font-medium text-gray-700 mb-1">Birthdate</label>
          <input v-model="form.birthdate" type="date" :min="minBirthdate" :max="maxBirthdate" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Contact No.</label>
          <input v-model="form.contact_no" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Civil Status</label>
          <select v-model="form.civil_status" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white">
            <option value="">Select Status</option>
            <option v-for="opt in CIVIL_STATUS_OPTIONS" :key="opt" :value="opt">
              {{ formatStatus(opt) }}
            </option>
          </select>
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Citizenship</label>
          <select v-model="form.citizenship" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white">
            <option value="">Select Citizenship</option>
            <option v-for="opt in CITIZENSHIP_OPTIONS" :key="opt" :value="opt">
              {{ formatStatus(opt) }}
            </option>
          </select>
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Inhabitant Type</label>
          <select v-model="form.inhabitant_type" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white">
            <option value="">Select Type</option>
            <option v-for="opt in INHABITANT_TYPE_OPTIONS" :key="opt" :value="opt">
              {{ formatStatus(opt) }}
            </option>
          </select>
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Birth Place</label>
          <input v-model="form.birth_place" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Sex</label>
          <select v-model="form.sex" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white">
            <option value="">Select Sex</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
          </select>
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Occupation</label>
          <input v-model="form.occupation" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Email Address</label>
          <input v-model="form.email_address" type="email" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1 sm:col-span-2">
          <label class="block text-sm font-medium text-gray-700 mb-1">Highest Educational Attainment</label>
          <input v-model="form.highest_educational_attainment" class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
        </div>
        <div class="col-span-1 sm:col-span-2">
          <h3 class="text-sm font-semibold text-gray-900 mb-2">Mother's Name</h3>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">First Name</label>
              <input v-model="form.mother_first_name" class="w-full text-xs p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">Middle Name</label>
              <input v-model="form.mother_middle_name" class="w-full text-xs p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">Last Name</label>
              <input v-model="form.mother_last_name" class="w-full text-xs p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none" />
            </div>
          </div>
        </div>
        <div class="col-span-1 sm:col-span-2">
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
import { 
  InhabitantService, type Inhabitant,
  CIVIL_STATUS_OPTIONS, CITIZENSHIP_OPTIONS, INHABITANT_TYPE_OPTIONS
} from '@/modules/inhabitants/services/inhabitantService';
import { extractErrorMessage } from '@/core/api';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';
import UiButton from '@/core/ui/components/UiButton.vue';

const router = useRouter();
const authStore = useAuthStore();
const isAdmin = computed(() => authStore.role === 'RoleAdmin');

// Birthdate constraints
const minBirthdate = computed(() => {
  const date = new Date();
  date.setFullYear(date.getFullYear() - 120);
  return date.toISOString().split('T')[0];
});

const maxBirthdate = computed(() => {
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
  birthdate: '',
  contact_no: '',
  address: '',
  civil_status: '',
  citizenship: '',
  inhabitant_type: '',
  sex: '',
  birth_place: '',
  occupation: '',
  email_address: '',
  highest_educational_attainment: '',
  mother_first_name: '',
  mother_middle_name: '',
  mother_last_name: ''
});

// Use inhabitants directly (server-side search is applied in loadInhabitants)
const filteredInhabitants = computed(() => inhabitants.value);

async function loadInhabitants() {
  loading.value = true;
  error.value = '';
  try {
    // Use search endpoint if there's a query, otherwise use getAll
    const response = searchQuery.value.trim()
      ? await InhabitantService.search(searchQuery.value, limit.value)
      : await InhabitantService.getAll(offset.value, limit.value);
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
    birthdate: '',
    contact_no: '',
    address: '',
    civil_status: '',
    citizenship: '',
    inhabitant_type: '',
    sex: '',
    birth_place: '',
    occupation: '',
    email_address: '',
    highest_educational_attainment: '',
    mother_first_name: '',
    mother_middle_name: '',
    mother_last_name: ''
  };
  formError.value = '';
  showCreateModal.value = true;
}

function validateBirthdate(birthdate: string): string | null {
  if (!birthdate || birthdate.trim() === '') {
    return null; // Empty is valid (optional field)
  }

  const birthdateDate = new Date(birthdate);
  const minDate = new Date();
  minDate.setFullYear(minDate.getFullYear() - 130);
  const maxDate = new Date();

  if (birthdateDate < minDate) {
    return 'Date out of range';
  }

  if (birthdateDate > maxDate) {
    return 'Birthdate cannot be in the future';
  }

  // Check for invalid dates like 0001-01-01
  if (birthdate.startsWith('0000') || birthdate.startsWith('0001')) {
    return 'Please enter a valid birthdate';
  }

  return null; // Valid
}

function validateEmail(email: string | undefined | null): string | null {
  if (!email || email.trim() === '') return null;
  const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
  if (!emailRegex.test(email)) {
    return 'Please enter a valid email address';
  }
  return null;
}

async function handleSubmit() {
  // Validate birthdate
  const birthdateError = validateBirthdate(form.value.birthdate);
  if (birthdateError) {
    formError.value = birthdateError;
    return;
  }

  const emailError = validateEmail(form.value.email_address);
  if (emailError) {
    formError.value = emailError;
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

function formatStatus(status: string | undefined | null): string {
  if (!status) return 'N/A';
  if (status === 'common_law_live_in') return 'Common Law/Live In';
  return status
    .split('_')
    .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
    .join(' ');
}

function formatBirthdate(birthdate: string | undefined | null): string {
  if (!birthdate || birthdate.trim() === '') {
    return 'N/A';
  }
  
  // Check for invalid dates like 0001-01-01, 0000-00-00, etc.
  const invalidDates = ['0001-01-01', '0000-00-00', '1900-01-01'];
  if (invalidDates.includes(birthdate)) {
    return 'N/A';
  }
  
  // Check if date starts with 0000 or 0001
  if (birthdate.startsWith('0000') || birthdate.startsWith('0001')) {
    return 'N/A';
  }
  
  return birthdate;
}

function viewDetails(inh: Inhabitant) {
  router.push({ name: 'inhabitant-detail', params: { id: inh.uuid!.toString() } });
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