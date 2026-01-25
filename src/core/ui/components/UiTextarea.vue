<template>
  <textarea
    :id="id"
    :name="name"
    :placeholder="placeholder"
    :disabled="disabled"
    :required="required"
    class="block w-full border border-gray-300 bg-white text-gray-900 rounded px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
    :rows="rows"
    :value="modelValue"
    @input="onInput"
  />
</template>

<script setup lang="ts">
const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  required: { type: Boolean, default: false },
  rows: { type: Number, default: 4 },
  name: { type: String, default: undefined },
  id: { type: String, default: undefined }
})

const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

function onInput(e: Event) {
  emit('update:modelValue', (e.target as HTMLTextAreaElement).value)
}
</script>
