<template>
  <input
    :id="id"
    :name="name"
    :type="type"
    :placeholder="placeholder"
    :disabled="disabled"
    :required="required"
    :autocomplete="autocomplete"
    class="block w-full border border-gray-300 bg-white text-gray-900 rounded px-3 py-2 focus:outline-none focus:ring focus:ring-blue-300"
    :value="modelValue"
    @input="onInput"
  />
</template>

<script setup lang="ts">
import type { PropType } from 'vue'

const props = defineProps({
  modelValue: { type: [String, Number] as PropType<string | number | null>, default: '' },
  type: { type: String, default: 'text' },
  placeholder: { type: String, default: '' },
  disabled: { type: Boolean, default: false },
  required: { type: Boolean, default: false },
  autocomplete: { type: String, default: 'off' },
  name: { type: String, default: undefined },
  id: { type: String, default: undefined }
})

const emit = defineEmits<{ (e: 'update:modelValue', value: string | number): void }>()

function onInput(e: Event) {
  const target = e.target as HTMLInputElement
  const value = props.type === 'number' ? Number(target.value) : target.value
  emit('update:modelValue', value)
}
</script>
