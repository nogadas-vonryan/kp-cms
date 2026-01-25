<template>
  <button
    type="button"
    role="switch"
    :aria-checked="modelValue"
    :disabled="disabled"
    class="inline-flex items-center border border-gray-300 rounded px-2 py-1 bg-white text-gray-900 select-none"
    @click="toggle"
  >
    <span
      class="inline-block h-4 w-8 border border-gray-300 rounded bg-white relative"
    >
      <span
        class="absolute top-0.5 left-0.5 h-3 w-3 rounded bg-gray-300 transition-transform"
        :class="modelValue ? 'translate-x-4 bg-blue-500' : 'translate-x-0'"
      />
    </span>
    <span v-if="label" class="ml-2">{{ label }}</span>
    <slot />
  </button>
</template>

<script setup lang="ts">
const props = defineProps({
  modelValue: { type: Boolean, default: false },
  label: { type: String, default: '' },
  disabled: { type: Boolean, default: false }
})

const emit = defineEmits<{ (e: 'update:modelValue', value: boolean): void }>()

function toggle() {
  if (props.disabled) return
  emit('update:modelValue', !props.modelValue)
}
</script>
