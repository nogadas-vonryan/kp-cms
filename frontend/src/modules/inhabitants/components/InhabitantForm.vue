<template>
  <div>
    <!-- Quick Mode Toggle -->
    <div v-if="!hideQuickModeToggle" class="mb-4 flex items-center gap-2">
      <button
        type="button"
        @click="isQuickMode = !isQuickMode"
        class="text-sm text-blue-600 hover:text-blue-700 hover:cursor-pointer underline"
      >
        {{ isQuickMode ? 'Show All Fields' : 'Quick Add (Name & Address Only)' }}
      </button>
    </div>

    <form @submit.prevent="handleSubmit" class="grid grid-cols-1 sm:grid-cols-2 gap-4">
      <!-- Required Fields (Always Visible) -->
      <div class="col-span-1">
        <label class="block text-sm font-medium text-gray-700 mb-1">
          First Name <span class="text-red-500">*</span>
        </label>
        <input
          v-model="formData.first_name"
          required
          class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
        />
      </div>
      
      <div class="col-span-1">
        <label class="block text-sm font-medium text-gray-700 mb-1">
          Last Name <span class="text-red-500">*</span>
        </label>
        <input
          v-model="formData.last_name"
          required
          class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
        />
      </div>

      <!-- Quick Mode: Address Only -->
      <div v-if="isQuickMode" class="col-span-1 sm:col-span-2">
        <label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
        <textarea
          v-model="formData.address"
          rows="2"
          class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
        ></textarea>
      </div>

      <!-- Full Mode: All Fields -->
      <template v-if="!isQuickMode">
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Middle Name</label>
          <input
            v-model="formData.middle_name"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Suffix</label>
          <input
            v-model="formData.suffix"
            placeholder="e.g. Jr., III"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Birthdate</label>
          <input
            v-model="formData.birthdate"
            type="date"
            :min="minBirthdate"
            :max="maxBirthdate"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Contact No.</label>
          <input
            v-model="formData.contact_no"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Civil Status</label>
          <select
            v-model="formData.civil_status"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white"
          >
            <option value="">Select Status</option>
            <option v-for="opt in CIVIL_STATUS_OPTIONS" :key="opt" :value="opt">
              {{ formatStatus(opt) }}
            </option>
          </select>
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Citizenship</label>
          <select
            v-model="formData.citizenship"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white"
          >
            <option value="">Select Citizenship</option>
            <option v-for="opt in CITIZENSHIP_OPTIONS" :key="opt" :value="opt">
              {{ formatStatus(opt) }}
            </option>
          </select>
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Inhabitant Type</label>
          <select
            v-model="formData.inhabitant_type"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white"
          >
            <option value="">Select Type</option>
            <option v-for="opt in INHABITANT_TYPE_OPTIONS" :key="opt" :value="opt">
              {{ formatStatus(opt) }}
            </option>
          </select>
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Birth Place</label>
          <input
            v-model="formData.birth_place"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Sex</label>
          <select
            v-model="formData.sex"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none bg-white"
          >
            <option value="">Select Sex</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
          </select>
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Occupation</label>
          <input
            v-model="formData.occupation"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1">
          <label class="block text-sm font-medium text-gray-700 mb-1">Email Address</label>
          <input
            v-model="formData.email_address"
            type="email"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1 sm:col-span-2">
          <label class="block text-sm font-medium text-gray-700 mb-1">
            Highest Educational Attainment
          </label>
          <input
            v-model="formData.highest_educational_attainment"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          />
        </div>
        
        <div class="col-span-1 sm:col-span-2">
          <h3 class="text-sm font-semibold text-gray-900 mb-2">Mother's Name</h3>
          <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">First Name</label>
              <input
                v-model="formData.mother_first_name"
                class="w-full text-xs p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">Middle Name</label>
              <input
                v-model="formData.mother_middle_name"
                class="w-full text-xs p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
              />
            </div>
            <div>
              <label class="block text-xs font-medium text-gray-500 mb-1">Last Name</label>
              <input
                v-model="formData.mother_last_name"
                class="w-full text-xs p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
              />
            </div>
          </div>
        </div>
        
        <div class="col-span-1 sm:col-span-2">
          <label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
          <textarea
            v-model="formData.address"
            rows="2"
            class="w-full text-sm p-2 border border-gray-300 rounded focus:ring-1 focus:ring-blue-500 outline-none"
          ></textarea>
        </div>
      </template>
    </form>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import {
  type Inhabitant,
  CIVIL_STATUS_OPTIONS,
  CITIZENSHIP_OPTIONS,
  INHABITANT_TYPE_OPTIONS
} from '@/modules/inhabitants/services/inhabitantService';

const props = withDefaults(
  defineProps<{
    initialData?: Partial<Inhabitant>;
    loading?: boolean;
    quickMode?: boolean;
    hideQuickModeToggle?: boolean;
  }>(),
  {
    loading: false,
    quickMode: false,
    hideQuickModeToggle: false
  }
);

const emit = defineEmits<{
  (e: 'submit', data: Inhabitant): void;
  (e: 'cancel'): void;
}>();

const isQuickMode = ref(props.quickMode);
const formData = ref<Inhabitant>({
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

// Date range for birthdate (120 years ago to today)
const minBirthdate = computed(() => {
  const date = new Date();
  date.setFullYear(date.getFullYear() - 120);
  return date.toISOString().split('T')[0];
});

const maxBirthdate = computed(() => {
  return new Date().toISOString().split('T')[0];
});

// Format status values for display
function formatStatus(value: string): string {
  return value
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
}

// Load initial data if provided
watch(
  () => props.initialData,
  (newData) => {
    if (newData) {
      formData.value = { ...formData.value, ...newData };
    }
  },
  { immediate: true }
);

// Expose submit method for parent
function handleSubmit() {
  emit('submit', formData.value);
}

// Expose form data for external access
defineExpose({ formData, handleSubmit });
</script>
