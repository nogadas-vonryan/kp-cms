<template>
  <select
    :id="id"
    :name="name"
    :disabled="disabled"
    :multiple="multiple"
    :size="multiple ? size : undefined"
    class="block w-full border border-gray-300 bg-white text-gray-900 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
    :value="multiple ? undefined : (modelValue ?? '')"
    @change="onChange"
  >
    <option v-if="placeholder" disabled value="">{{ placeholder }}</option>
    <option
      v-for="opt in options" 
      :key="String(opt.value)" 
      :value="opt.value"
      :selected="multiple ? isSelected(opt.value) : undefined"
    >
      {{ opt.label }}
    </option>
    <option v-if="options.length === 0" disabled value="">No options available</option>
    <slot />
  </select>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'

type SelectValue = string | number
type Option = { label: string; value: SelectValue }

const props = defineProps({
  modelValue: { type: [String, Number, Array] as PropType<SelectValue | SelectValue[] | null>, default: null },
  options: { type: Array as PropType<Option[]>, default: () => [] },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  name: { type: String, default: undefined },
  id: { type: String, default: undefined },
  multiple: { type: Boolean, default: false },
  size: { type: Number, default: 6 }
})

const emit = defineEmits<{ (e: 'update:modelValue', value: SelectValue | SelectValue[]): void }>()

function onChange(e: Event) {
  const target = e.target as HTMLSelectElement
  if (props.multiple) {
    const selected = Array.from(target.selectedOptions).map((opt) => normalizeValue(opt.value))
    emit('update:modelValue', selected)
    return
  }

  const raw = target.value
  emit('update:modelValue', normalizeValue(raw))
}

function normalizeValue(raw: string): SelectValue {
  const match = props.options.find((opt) => String(opt.value) === raw)
  return match ? match.value : raw
}

function isSelected(value: SelectValue) {
  if (!Array.isArray(props.modelValue)) return false
  return props.modelValue.some((v) => String(v) === String(value))
}
</script>
