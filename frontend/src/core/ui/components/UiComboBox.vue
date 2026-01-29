<template>
  <div class="relative w-full" v-click-outside="close">
    <UiInput
      v-model="internalValue"
      :placeholder="placeholder"
      @focus="showDropdown = true"
      @input="showDropdown = true"
    />
    
    <div 
      v-if="showDropdown && filteredOptions.length > 0" 
      class="absolute z-10 w-full mt-1 bg-white border border-gray-300 rounded shadow-sm max-h-48 overflow-y-auto"
    >
      <div
        v-for="option in filteredOptions"
        :key="option"
        @click="selectOption(option)"
        class="px-3 py-2 text-sm text-gray-900 cursor-pointer hover:bg-gray-50 border-b border-gray-100 last:border-0"
      >
        {{ option }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import UiInput from './UiInput.vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  options: { type: Array as () => string[], default: () => [] },
  placeholder: { type: String, default: '' }
})

const emit = defineEmits(['update:modelValue', 'option-selected'])
const showDropdown = ref(false)
const internalValue = ref(props.modelValue)

watch(() => props.modelValue, (newVal) => internalValue.value = newVal)
watch(internalValue, (newVal) => emit('update:modelValue', newVal))

const filteredOptions = computed(() => {
  if (!internalValue.value) return props.options
  return props.options.filter(opt => 
    opt.toLowerCase().includes(internalValue.value.toLowerCase())
  )
})

const selectOption = (opt: string) => {
  internalValue.value = opt
  showDropdown.value = false
  emit('option-selected', opt)
}

const close = () => showDropdown.value = false

// Simple directive for clicking outside
const vClickOutside = {
  mounted(el: any, binding: any) {
    el.clickOutsideEvent = (event: Event) => {
      if (!(el === event.target || el.contains(event.target))) {
        binding.value()
      }
    }
    document.addEventListener('click', el.clickOutsideEvent)
  },
  unmounted(el: any) {
    document.removeEventListener('click', el.clickOutsideEvent)
  }
}
</script>