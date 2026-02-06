<template>
  <div class="space-y-4">
    <div class="flex items-center gap-2 text-xs sm:text-sm text-gray-600 min-w-0">
      <router-link to="/inhabitants" class="hover:text-gray-900 truncate">Inhabitants</router-link>
      <span class="shrink-0">/</span>
      <span class="text-gray-900 font-medium wrap-break-word">{{ inhabitant ? getFullName(inhabitant) : 'Loading...' }}</span>
    </div>

    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
      <div class="min-w-0">
        <h1 class="text-2xl sm:text-3xl font-bold text-gray-900 wrap-break-word">
          {{ inhabitant ? getFullName(inhabitant) : 'Inhabitant Details' }}
        </h1>
        <p v-if="inhabitant" class="text-xs sm:text-sm text-gray-600 mt-1">
          ID: <span class="font-mono">{{ inhabitant.uuid }}</span>
        </p>
      </div>
      <div v-if="isAdmin && inhabitant" class="flex gap-2 shrink-0 ml-auto">
        <button
          @click="openEditModal"
          title="Edit"
          class="p-2 text-gray-600 hover:text-gray-900 hover:bg-gray-100 rounded border border-gray-300 transition-colors"
        >
          <Edit3 :size="20" :stroke-width="1.4" />
        </button>
        
        <button 
          @click="showDeleteModal = true" 
          title="Delete" 
          class="p-2 text-red-600 hover:text-red-900 hover:bg-red-50 rounded border border-red-300 transition-colors"
        >
          <Trash2 :size="20" :stroke-width="1.4" />
        </button>
      </div>
    </div>

    <div v-if="loading" class="p-8 text-center text-gray-600">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-2"></div>
      Loading inhabitant details...
    </div>

    <UiAlert v-if="error" type="error" class="mb-4">
      {{ error }}
    </UiAlert>

    <template v-if="!loading && inhabitant">
      <UiCard>
        <div class="space-y-4">
          <h2 class="text-lg font-semibold text-gray-900">Profile Information</h2>
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-x-8 gap-y-4">
            <div class="min-w-0">
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Full Name</label>
              <p class="text-sm text-gray-900 font-medium wrap-break-word">{{ getFullName(inhabitant) || 'N/A' }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Birthdate</label>
              <p class="text-sm text-gray-900">{{ formatBirthdate(inhabitant.birthdate) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Contact Number</label>
              <p class="text-sm text-gray-900">{{ formatField(inhabitant.contact_no) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Sex</label>
              <p class="text-sm text-gray-900">{{ formatField(inhabitant.sex) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Civil Status</label>
              <p class="text-sm text-gray-900">{{ formatStatus(inhabitant.civil_status) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Citizenship</label>
              <p class="text-sm text-gray-900">{{ formatStatus(inhabitant.citizenship) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Inhabitant Type</label>
              <p class="text-sm text-gray-900">{{ formatStatus(inhabitant.inhabitant_type) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Birth Place</label>
              <p class="text-sm text-gray-900">{{ formatField(inhabitant.birth_place) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Occupation</label>
              <p class="text-sm text-gray-900">{{ formatField(inhabitant.occupation) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Email Address</label>
              <p class="text-sm text-gray-900 truncate">{{ formatField(inhabitant.email_address) }}</p>
            </div>
            <div class="sm:col-span-2">
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Highest Educational Attainment</label>
              <p class="text-sm text-gray-900">{{ formatField(inhabitant.highest_educational_attainment) }}</p>
            </div>
            <div class="sm:col-span-2">
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Mother's Name</label>
              <p class="text-sm text-gray-900">{{ getMotherFullName(inhabitant) || 'N/A' }}</p>
            </div>
            <div class="sm:col-span-2">
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Address</label>
              <p class="text-sm text-gray-900 wrap-break-word">{{ formatField(inhabitant.address) }}</p>
            </div>
          </div>
        </div>
      </UiCard>

      <div class="space-y-3">
        <h2 class="text-lg font-semibold text-gray-900 px-1">Related Documents</h2>

        <div v-if="documentsLoading" class="p-8 text-center text-gray-600 bg-white rounded-lg border border-gray-200">
          <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mx-auto mb-2"></div>
          Searching for documents...
        </div>

        <UiAlert v-else-if="documentsError" type="error">{{ documentsError }}</UiAlert>

        <div v-else-if="documents.length > 0" class="grid gap-3">
          <div 
            v-for="doc in documents" 
            :key="doc.uuid"
            @click="navigateToDocument(doc.uuid)"
            class="p-4 bg-white border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-blue-300 transition-all cursor-pointer group"
          >
            <div class="flex items-center justify-between gap-4">
              <div class="flex-1 min-w-0">
                <h3 class="font-medium text-gray-900 text-sm group-hover:text-blue-700 transition-colors truncate">
                  {{ doc.title }}
                </h3>
                <div class="flex flex-wrap items-center gap-y-1 text-xs text-gray-500 mt-1">
                  <span class="font-mono bg-gray-100 px-1 rounded">{{ doc.code || 'N/A' }}</span>
                  <span class="mx-2 text-gray-300">•</span>
                  <span class="truncate">{{ doc.folder_name }}</span>
                </div>
                
                <div v-if="doc.fields?.complainants || doc.fields?.respondents" class="text-[11px] text-gray-500 mt-2 space-y-0.5 border-t border-gray-50 pt-2">
                  <p v-if="doc.fields.complainants?.length" class="truncate">
                    <span class="font-medium">Complainants:</span> {{ doc.fields.complainants.join(', ') }}
                  </p>
                  <p v-if="doc.fields.respondents?.length" class="truncate">
                    <span class="font-medium">Respondents:</span> {{ doc.fields.respondents.join(', ') }}
                  </p>
                </div>
              </div>
              <ChevronRight :size="18" class="text-gray-400 shrink-0 group-hover:translate-x-1 transition-transform" />
            </div>
          </div>
        </div>

        <div v-else class="text-center py-12 bg-white rounded-lg border border-dashed border-gray-300 text-gray-500">
          <p class="text-sm font-medium">No documents found</p>
          <p class="text-xs mt-1">This inhabitant is not linked to any document records.</p>
        </div>
      </div>
    </template>

    <UiModal v-model:open="showEditModal" title="Edit Inhabitant">
      <form @submit.prevent="handleSubmit" class="grid grid-cols-1 sm:grid-cols-2 gap-4">
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">First Name <span class="text-red-500">*</span></label>
          <UiInput v-model="form.first_name" required class="w-full" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Last Name <span class="text-red-500">*</span></label>
          <UiInput v-model="form.last_name" required class="w-full" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Middle Name</label>
          <UiInput v-model="form.middle_name" class="w-full" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Suffix</label>
          <UiInput v-model="form.suffix" placeholder="e.g. Jr., III" class="w-full" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Birthdate</label>
          <UiInput v-model="form.birthdate" type="date" :min="minBirthdate" :max="maxBirthdate" class="w-full" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Contact No.</label>
          <UiInput v-model="form.contact_no" class="w-full" />
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
          <UiInput v-model="form.birth_place" class="w-full" />
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
          <UiInput v-model="form.occupation" class="w-full" />
        </div>
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Email Address</label>
          <UiInput v-model="form.email_address" type="email" class="w-full" />
        </div>
        <div class="col-span-1 sm:col-span-2">
          <label class="block text-sm font-medium text-gray-700 mb-1">Highest Educational Attainment</label>
          <UiInput v-model="form.highest_educational_attainment" class="w-full" />
        </div>
        <div class="col-span-1 sm:col-span-2 mt-2">
          <h3 class="text-sm font-semibold text-gray-900 mb-2">Mother's Name</h3>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">First Name</label>
              <UiInput v-model="form.mother_first_name" class="w-full" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">Middle Name</label>
              <UiInput v-model="form.mother_middle_name" class="w-full" />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">Last Name</label>
              <UiInput v-model="form.mother_last_name" class="w-full" />
            </div>
          </div>
        </div>
        <div class="col-span-1 sm:col-span-2">
          <label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
          <textarea 
            v-model="form.address" 
            rows="3" 
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none transition-shadow"
          ></textarea>
        </div>
      </form>

      <template #footer>
        <div class="flex w-full flex-col gap-3">
          <UiAlert v-if="formError" type="error">{{ formError }}</UiAlert>
          <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
            <UiButton @click="showEditModal = false" variant="secondary" class="w-full sm:w-auto">Cancel</UiButton>
            <UiButton @click="handleSubmit" :loading="submitting" class="w-full sm:w-auto">
              Save Changes
            </UiButton>
          </div>
        </div>
      </template>
    </UiModal>

    <UiModal v-model:open="showDeleteModal" title="Delete Inhabitant" :prevent-close="deleting">
      <p class="text-gray-700 wrap-break-word mb-2">
        Are you sure you want to delete "<strong>{{ inhabitant ? getFullName(inhabitant) : 'this inhabitant' }}</strong>"?
      </p>
      
      <div class="space-y-4">
        <div class="bg-red-50 p-3 rounded border border-red-100">
          <p class="text-sm text-red-800 font-medium">This action is permanent.</p>
          <p class="text-sm text-red-700 mt-1 mb-2">
              To confirm, please type <span class="font-mono font-bold">"{{ DELETE_PHRASE.toLowerCase() }}"</span> below:
          </p>
          <UiInput 
            v-model="deleteConfirmationInput" 
            placeholder="Type the confirmation phrase"
            @keyup.enter="canDelete && !deleting && handleDelete()"
            class="w-full"
          />
        </div>

        <UiAlert v-if="deleteError" type="error">
          {{ deleteError }}
        </UiAlert>
      </div>

      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <UiButton :disabled="deleting" @click="showDeleteModal = false" variant="secondary" class="w-full sm:w-auto">Cancel</UiButton>
          <UiButton 
            @click="handleDelete" 
            :loading="deleting" 
            :disabled="!canDelete" 
            variant="danger"
            class="w-full sm:w-auto"
          >
            Delete Permanently
          </UiButton>
        </div>
      </template>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { 
  Edit3, Trash2, ChevronRight
} from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { 
  InhabitantService, type Inhabitant,
  CIVIL_STATUS_OPTIONS, CITIZENSHIP_OPTIONS, INHABITANT_TYPE_OPTIONS
} from '@/modules/inhabitants/services/inhabitantService';
import type { Document } from '@/types';
import { extractErrorMessage } from '@/core/api';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiButton from '@/core/ui/components/UiButton.vue';

const route = useRoute();
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

// State
const inhabitant = ref<Inhabitant | null>(null);
const documents = ref<Document[]>([]);
const loading = ref(false);
const documentsLoading = ref(false);
const error = ref('');
const documentsError = ref('');

// Edit Modal State
const showEditModal = ref(false);
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

// Delete Modal State
const showDeleteModal = ref(false);
const deleting = ref(false);
const deleteError = ref('');
const deleteConfirmationInput = ref('');
const DELETE_PHRASE = 'i want to delete it';

const canDelete = computed(() => {
  return deleteConfirmationInput.value.toLowerCase() === DELETE_PHRASE;
});

watch(showDeleteModal, (isOpen) => {
  if (!isOpen) {
    deleteConfirmationInput.value = '';
    deleteError.value = '';
  }
});

onMounted(() => {
  loadInhabitantDetails();
});

async function loadInhabitantDetails() {
  const id = route.params.id as string;
  if (!id) {
    error.value = 'Invalid inhabitant ID';
    return;
  }

  loading.value = true;
  error.value = '';
  
  try {
    const response = await InhabitantService.getById(id);
    inhabitant.value = response.data;
    await loadRelatedDocuments(id);
  } catch (err: any) {
    error.value = extractErrorMessage(err) || 'Failed to load inhabitant details';
  } finally {
    loading.value = false;
  }
}

async function loadRelatedDocuments(id: string) {
  documentsLoading.value = true;
  documentsError.value = '';
  
  try {
    const response = await InhabitantService.getInhabitantDocuments(id);
    documents.value = response.data || [];
  } catch (err: any) {
    documentsError.value = extractErrorMessage(err) || 'Failed to load related documents';
  } finally {
    documentsLoading.value = false;
  }
}

function getFullName(inh: Inhabitant): string {
  const parts = [
    inh.first_name,
    inh.middle_name,
    inh.last_name,
    inh.suffix
  ].filter(Boolean);
  return parts.join(' ');
}

function getMotherFullName(inh: Inhabitant): string {
  const parts = [
    inh.mother_first_name,
    inh.mother_middle_name,
    inh.mother_last_name
  ].filter(Boolean);
  return parts.join(' ');
}

function formatField(value: string | undefined | null): string {
  if (!value || value.trim() === '') return 'N/A';
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

/**
 * Checks if a date string is one of the common "Zero" or placeholder dates
 * used by databases like MySQL or Go.
 */
function isInvalidDate(birthdate: string | undefined | null): boolean {
  if (!birthdate || birthdate.trim() === '') return true;
  const invalidPatterns = ['0001-01-01', '0000-00-00', '1900-01-01'];
  return (
    invalidPatterns.includes(birthdate) || 
    birthdate.startsWith('0000') || 
    birthdate.startsWith('0001')
  );
}

function formatBirthdate(birthdate: string | undefined | null): string {
  if (isInvalidDate(birthdate)) {
    return 'N/A';
  }
  return birthdate!;
}

function navigateToDocument(uuid: string) {
  router.push({ name: 'document-detail', params: { documentId: uuid } });
}

function openEditModal() {
  if (inhabitant.value) {
    const birthdate = inhabitant.value.birthdate;
    
    form.value = { 
      ...inhabitant.value,
      // If the existing birthdate is invalid/placeholder, 
      // clear it so the date picker doesn't show 0001-01-01
      birthdate: isInvalidDate(birthdate) ? '' : birthdate
    };
    
    formError.value = '';
    showEditModal.value = true;
  }
}

function validateBirthdate(birthdate: string): string | null {
  if (!birthdate || birthdate.trim() === '' || isInvalidDate(birthdate)) return null;
  
  const birthdateDate = new Date(birthdate);
  const minDate = new Date();
  minDate.setFullYear(minDate.getFullYear() - 130);
  const maxDate = new Date();

  if (birthdateDate < minDate) return 'Date out of range';
  if (birthdateDate > maxDate) return 'Birthdate cannot be in the future';
  return null;
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
  if (!inhabitant.value?.uuid) return;
  
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
    await InhabitantService.update(inhabitant.value.uuid!, form.value);
    showEditModal.value = false;
    await loadInhabitantDetails();
  } catch (err: any) {
    formError.value = extractErrorMessage(err) || 'Failed to update inhabitant';
  } finally {
    submitting.value = false;
  }
}

async function handleDelete() {
  if (!inhabitant.value || !canDelete.value) return;
  deleting.value = true;
  deleteError.value = '';
  try {
    await InhabitantService.delete(inhabitant.value.uuid!);
    showDeleteModal.value = false;
    router.push({ name: 'inhabitants' });
  } catch (err: any) {
    deleteError.value = extractErrorMessage(err) || 'Failed to delete inhabitant';
  } finally {
    deleting.value = false;
  }
}
</script>