<template>
  <label class="inline-flex items-center gap-2 select-none">
    <input
      type="checkbox"
      :id="id"
      :name="name"
      :disabled="disabled"
      class="h-4 w-4 border border-gray-300 rounded bg-white text-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
      :checked="modelValue"
      @change="onChange"
    />
    <span v-if="label" class="text-gray-900">{{ label }}</span>
    <slot />
  </label>
</template>

<script setup lang="ts">
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  label: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  name: { type: String, default: undefined },
  id: { type: String, default: undefined }
})

const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()

function onChange(e: Event) {
  emit('update:modelValue', (e.target as HTMLInputElement).checked)
}
</script>
