<template>
  <div class="space-y-2">
    <!-- Selected Inhabitants -->
    <div v-if="modelValue.length > 0" class="space-y-2">
      <div 
        v-for="inhabitant in selectedInhabitants" 
        :key="inhabitant.id"
        class="flex items-center justify-between gap-2 p-2 bg-gray-50 border border-gray-200 rounded"
      >
        <span class="text-sm text-gray-900">{{ formatInhabitantName(inhabitant) }}</span>
        <button 
          type="button"
          @click="removeInhabitant(inhabitant.id!)"
          class="p-1 text-red-600 hover:bg-red-100 rounded transition-colors shrink-0"
          title="Remove"
        >
          <XIcon :size="14" />
        </button>
      </div>
    </div>

    <!-- Search & Add -->
    <div class="relative">
      <div class="relative">
        <input
          v-model="searchQuery"
          @input="handleSearch"
          @focus="showDropdown = true"
          type="text"
          placeholder="Search inhabitants..."
          class="w-full text-sm p-2 pl-8 border border-gray-300 rounded bg-white focus:ring-1 focus:ring-blue-500 outline-none"
        />
        <SearchIcon class="absolute left-2.5 top-1/2 -translate-y-1/2 text-gray-400" :size="14" />
      </div>

      <!-- Dropdown Results -->
      <div 
        v-if="showDropdown && (searchResults.length > 0 || searching)"
        class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded-lg shadow-lg max-h-48 overflow-y-auto"
      >
        <div v-if="searching" class="p-3 text-sm text-gray-600 text-center">
          Searching...
        </div>
        <div 
          v-for="inhabitant in availableInhabitants" 
          :key="inhabitant.id"
          @click="addInhabitant(inhabitant)"
          class="p-2 hover:bg-gray-50 cursor-pointer border-b border-gray-100 last:border-b-0"
        >
          <div class="text-sm text-gray-900">{{ formatInhabitantName(inhabitant) }}</div>
          <div class="text-xs text-gray-500">
            {{ inhabitant.address || 'No address' }}
          </div>
        </div>
        <div v-if="!searching && searchResults.length === 0" class="p-3 text-sm text-gray-600 text-center">
          No inhabitants found
        </div>
      </div>
    </div>

    <!-- Quick Add Button -->
    <button
      v-if="!hideAddButton"
      type="button"
      @click="showCreateModal = true"
      class="w-full text-sm p-2 border border-dashed border-gray-300 rounded text-gray-600 hover:text-gray-900 hover:border-gray-400 transition-colors flex items-center justify-center gap-2"
    >
      <PlusIcon :size="14" />
      <span>Create New Inhabitant</span>
    </button>

    <!-- Create Inhabitant Modal -->
    <InhabitantCreateModal
      v-model:open="showCreateModal"
      @created="handleInhabitantCreated"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onUnmounted } from 'vue';
import { X as XIcon, Search as SearchIcon, Plus as PlusIcon } from 'lucide-vue-next';
import { InhabitantService, type Inhabitant } from '@/modules/inhabitants/services/inhabitantService';
import InhabitantCreateModal from '@/modules/inhabitants/components/InhabitantCreateModal.vue';

const props = withDefaults(defineProps<{
  modelValue: number[];
  hideAddButton?: boolean;
}>(), {
  hideAddButton: false
});

const emit = defineEmits<{
  (e: 'update:modelValue', value: number[]): void;
}>();

const searchQuery = ref('');
const searchResults = ref<Inhabitant[]>([]);
const selectedInhabitants = ref<Inhabitant[]>([]);
const showDropdown = ref(false);
const searching = ref(false);
const showCreateModal = ref(false);
let searchTimeout: ReturnType<typeof setTimeout> | null = null;

// Filter out already selected inhabitants from search results
const availableInhabitants = computed(() => {
  return searchResults.value.filter(
    inhabitant => !props.modelValue.includes(inhabitant.id!)
  );
});

// Format inhabitant name for display
function formatInhabitantName(inhabitant: Inhabitant): string {
  const parts = [
    inhabitant.first_name,
    inhabitant.middle_name,
    inhabitant.last_name,
    inhabitant.suffix
  ].filter(Boolean);
  return parts.join(' ') || 'Unnamed';
}

// Load details for selected inhabitants
async function loadSelectedInhabitants() {
  if (props.modelValue.length === 0) {
    selectedInhabitants.value = [];
    return;
  }

  try {
    const promises = props.modelValue.map(id => InhabitantService.getById(id));
    const responses = await Promise.all(promises);
    selectedInhabitants.value = responses.map(r => r.data);
  } catch (error) {
    console.error('Failed to load selected inhabitants:', error);
  }
}

// Search inhabitants with debounce
function handleSearch() {
  if (searchTimeout) {
    clearTimeout(searchTimeout);
  }

  if (!searchQuery.value.trim()) {
    searchResults.value = [];
    showDropdown.value = false;
    return;
  }

  searching.value = true;
  searchTimeout = setTimeout(async () => {
    try {
      const response = await InhabitantService.search(searchQuery.value, 20);
      searchResults.value = response.data;
      showDropdown.value = true;
    } catch (error) {
      console.error('Search failed:', error);
      searchResults.value = [];
    } finally {
      searching.value = false;
    }
  }, 300);
}

// Add inhabitant to selection
function addInhabitant(inhabitant: Inhabitant) {
  if (!inhabitant.id || props.modelValue.includes(inhabitant.id)) return;

  const newValue = [...props.modelValue, inhabitant.id];
  emit('update:modelValue', newValue);
  
  selectedInhabitants.value.push(inhabitant);
  searchQuery.value = '';
  searchResults.value = [];
  showDropdown.value = false;
}

// Remove inhabitant from selection
function removeInhabitant(id: number) {
  const newValue = props.modelValue.filter(itemId => itemId !== id);
  emit('update:modelValue', newValue);
  
  selectedInhabitants.value = selectedInhabitants.value.filter(i => i.id !== id);
}

// Handle new inhabitant creation - auto-select it
function handleInhabitantCreated(inhabitant: Inhabitant) {
  if (!inhabitant.id) return;
  
  // Add to model value
  const newValue = [...props.modelValue, inhabitant.id];
  emit('update:modelValue', newValue);
  
  // Add to selected list
  selectedInhabitants.value.push(inhabitant);
}

// Close dropdown when clicking outside
function handleClickOutside(event: MouseEvent) {
  const target = event.target as HTMLElement;
  if (!target.closest('.relative')) {
    showDropdown.value = false;
  }
}

// Load selected inhabitants when IDs change
watch(() => props.modelValue, () => {
  loadSelectedInhabitants();
}, { immediate: true });

// Add click outside listener
if (typeof window !== 'undefined') {
  document.addEventListener('click', handleClickOutside);
}

onUnmounted(() => {
  if (typeof window !== 'undefined') {
    document.removeEventListener('click', handleClickOutside);
  }
});
</script>
