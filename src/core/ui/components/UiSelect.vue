<template>
  <select
    :id="id"
    :name="name"
    :disabled="disabled"
    class="block w-full border border-gray-300 bg-white text-gray-900 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
    :value="modelValue"
    @change="onChange"
  >
    <option v-if="placeholder" disabled value="">{{ placeholder }}</option>
    <option v-for="opt in options" :key="String(opt.value)" :value="opt.value">{{ opt.label }}</option>
    <slot />
  </select>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'

type Option = { label: string; value: string | number }

const props = defineProps({
  modelValue: { type: [String, Number] as PropType<string | number | null>, default: '' },
  options: { type: Array as PropType<Option[]>, default: () => [] },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  name: { type: String, default: undefined },
  id: { type: String, default: undefined }
})

const emit = defineEmits<{ (e: 'update:modelValue', value: string | number): void }>()

function onChange(e: Event) {
  const target = e.target as HTMLSelectElement
  const raw = target.value
  const num = Number(raw)
  emit('update:modelValue', isNaN(num) ? raw : num)
}
</script>
