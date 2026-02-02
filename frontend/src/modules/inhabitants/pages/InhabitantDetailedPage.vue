<template>
  <div class="space-y-4">
    <!-- Breadcrumb -->
    <div class="flex items-center gap-2 text-xs sm:text-sm text-gray-600 min-w-0">
      <router-link to="/inhabitants" class="hover:text-gray-900 truncate">Inhabitants</router-link>
      <span class="shrink-0">/</span>
      <span class="text-gray-900 font-medium wrap-break-word">{{ inhabitant ? getFullName(inhabitant) : 'Loading...' }}</span>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="flex justify-center items-center min-h-96">
      <div class="text-center space-y-2">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
        <p class="text-sm text-gray-600">Loading inhabitant details...</p>
      </div>
    </div>

    <!-- Error State -->
    <UiSystemNotice v-if="error" type="error" label="Error" :modelValue="true">
      <template #title>
        <span class="px-2 text-sm text-gray-700 leading-relaxed">{{ error }}</span>
      </template>
    </UiSystemNotice>

    <!-- Content -->
    <template v-if="!loading && inhabitant">
      <!-- Profile Card -->
      <UiCard>
        <div class="space-y-4">
          <div class="flex items-start justify-between">
            <h2 class="text-lg font-semibold text-gray-900">Profile Information</h2>
            <div v-if="isAdmin" class="flex gap-2">
              <UiButton @click="openEditModal" variant="secondary" class="flex items-center gap-2">
                <Edit3 :size="16" />
                <span>Edit</span>
              </UiButton>
              <UiButton @click="showDeleteModal = true" class="flex items-center gap-2 bg-red-600 hover:bg-red-700">
                <Trash2 :size="16" />
                <span>Delete</span>
              </UiButton>
            </div>
          </div>

          <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Full Name</label>
              <p class="text-sm text-gray-900">{{ getFullName(inhabitant) || 'N/A' }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Birthday</label>
              <p class="text-sm text-gray-900">{{ formatBirthday(inhabitant.birthday) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Contact Number</label>
              <p class="text-sm text-gray-900">{{ formatField(inhabitant.contact_no) }}</p>
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 uppercase mb-1">Address</label>
              <p class="text-sm text-gray-900">{{ formatField(inhabitant.address) }}</p>
            </div>
          </div>
        </div>
      </UiCard>

      <!-- Related Documents Section -->
      <UiCard>
        <div class="space-y-4">
          <h2 class="text-lg font-semibold text-gray-900">Related Documents</h2>

          <!-- Documents Loading State -->
          <div v-if="documentsLoading" class="flex justify-center items-center py-8">
            <div class="text-center space-y-2">
              <div class="animate-spin rounded-full h-6 w-6 border-b-2 border-blue-600 mx-auto"></div>
              <p class="text-xs text-gray-600">Searching for related documents...</p>
            </div>
          </div>

          <!-- Documents Error State -->
          <UiAlert v-else-if="documentsError" type="error">{{ documentsError }}</UiAlert>

          <!-- Documents List -->
          <div v-else-if="documents.length > 0" class="space-y-2">
            <div 
              v-for="doc in documents" 
              :key="doc.uuid"
              @click="navigateToDocument(doc.uuid)"
              class="p-4 border border-gray-200 rounded-lg hover:bg-gray-50 hover:border-blue-300 transition-all cursor-pointer"
            >
              <div class="flex items-start justify-between gap-2">
                <div class="flex-1 min-w-0">
                  <h3 class="font-medium text-gray-900 text-sm">{{ doc.title }}</h3>
                  <p class="text-xs text-gray-600 mt-1">
                    <span class="font-medium">Code:</span> {{ doc.code || 'N/A' }}
                    <span class="mx-2">•</span>
                    <span class="font-medium">Folder:</span> {{ doc.folder_name }}
                  </p>
                  <div v-if="doc.fields?.complainants || doc.fields?.respondents" class="text-xs text-gray-500 mt-2 space-y-1">
                    <p v-if="doc.fields.complainants && doc.fields.complainants.length">
                      <span class="font-medium">Complainants:</span> {{ doc.fields.complainants.join(', ') }}
                    </p>
                    <p v-if="doc.fields.respondents && doc.fields.respondents.length">
                      <span class="font-medium">Respondents:</span> {{ doc.fields.respondents.join(', ') }}
                    </p>
                  </div>
                </div>
                <ChevronRight :size="16" class="text-gray-400 shrink-0 mt-1" />
              </div>
            </div>
          </div>

          <!-- Empty State -->
          <div v-else class="text-center py-8 text-gray-600">
            <p class="text-sm">No documents found for this inhabitant</p>
          </div>
        </div>
      </UiCard>
    </template>

    <!-- Edit Modal -->
    <UiModal v-model:open="showEditModal" title="Edit Inhabitant">
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
            <UiButton @click="showEditModal = false">Cancel</UiButton>
            <UiButton @click="handleSubmit" :disabled="submitting" variant="secondary" class="flex items-center gap-2">
              <Check :size="16" />
              <span>{{ submitting ? 'Saving...' : 'Save' }}</span>
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
        <div class="bg-red-50 p-4 rounded-lg border border-red-200 space-y-3">
          <div class="space-y-2">
            <p class="text-sm text-red-800 font-medium">This action is permanent.</p>
            <p class="text-sm text-red-700 mt-1 mb-2">
                To confirm, please type <span class="font-mono font-bold">"I want to delete it"</span> below:
            </p>
          </div>

          <UiInput 
            v-model="deleteConfirmationInput" 
            placeholder="Type the confirmation phrase"
            class="bg-white"
            @keyup.enter="canDelete && !deleting && handleDelete()"
          />
        </div>

        <UiAlert v-if="deleteError" type="error" class="text-xs">
          {{ deleteError }}
        </UiAlert>
      </div>

      <template #footer>
        <div class="flex w-full flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <UiButton :disabled="deleting" @click="showDeleteModal = false" variant="secondary">Cancel</UiButton>
          <UiButton 
            @click="handleDelete" 
            :disabled="!canDelete || deleting" 
            variant="danger"
            class="flex items-center gap-2"
          >
            <Trash2 :size="16" />
            <span>{{ deleting ? 'Deleting...' : 'Confirm Delete' }}</span>
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
  Edit3, Trash2, Check, ChevronRight
} from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { InhabitantService, type Inhabitant } from '@/modules/inhabitants/services/inhabitantService';
import type { Document } from '@/types';
import { extractErrorMessage } from '@/core/api';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiModal from '@/core/ui/components/UiModal.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiSystemNotice from '@/core/ui/components/UiSystemNotice.vue';
import UiButton from '@/core/ui/components/UiButton.vue';

const route = useRoute();
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
  birthday: '',
  contact_no: '',
  address: ''
});

// Delete Modal State
const showDeleteModal = ref(false);
const deleting = ref(false);
const deleteError = ref('');
const deleteConfirmationInput = ref('');
const DELETE_PHRASE = 'I want to delete it';

const canDelete = computed(() => {
  return deleteConfirmationInput.value.toLowerCase() === DELETE_PHRASE.toLowerCase();
});

// Reset confirmation input when modal closes
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
  const id = parseInt(route.params.id as string);
  if (isNaN(id)) {
    error.value = 'Invalid inhabitant ID';
    return;
  }

  loading.value = true;
  error.value = '';
  
  try {
    const response = await InhabitantService.getById(id);
    inhabitant.value = response.data;
    
    // Load related documents
    await loadRelatedDocuments(id);
  } catch (err: any) {
    error.value = extractErrorMessage(err) || 'Failed to load inhabitant details';
  } finally {
    loading.value = false;
  }
}

async function loadRelatedDocuments(id: number) {
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

function navigateToDocument(uuid: string) {
  router.push({ name: 'document-detail', params: { documentId: uuid } });
}

function openEditModal() {
  if (inhabitant.value) {
    form.value = { ...inhabitant.value };
    formError.value = '';
    showEditModal.value = true;
  }
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
  if (!inhabitant.value?.id) return;

  // Validate birthday
  const birthdayError = validateBirthday(form.value.birthday);
  if (birthdayError) {
    formError.value = birthdayError;
    return;
  }

  submitting.value = true;
  formError.value = '';
  
  try {
    await InhabitantService.update(inhabitant.value.id, form.value);
    showEditModal.value = false;
    
    // Reload details and documents (name might have changed)
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
    await InhabitantService.delete(inhabitant.value.id!);
    showDeleteModal.value = false;
    router.push({ name: 'inhabitants' });
  } catch (err: any) {
    deleteError.value = extractErrorMessage(err) || 'Failed to delete inhabitant';
  } finally {
    deleting.value = false;
  }
}
</script>
