<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    class="inline-flex items-center justify-center whitespace-nowrap border border-gray-300 bg-white text-gray-900 rounded px-3 py-2 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
    :class="[
      block ? 'w-full' : 'w-auto'
    ]"
    @click="onClick"
  >
    <span v-if="loading" class="mr-2 h-3 w-3 rounded-full border-2 border-gray-300 border-t-transparent animate-spin"></span>
    <slot />
  </button>
</template>

<script setup lang="ts">
const props = defineProps({
  type: { type: String as () => 'button' | 'submit' | 'reset', default: 'button' },
  disabled: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  block: { type: Boolean, default: false }
})

const emit = defineEmits<{ (e: 'click', evt: MouseEvent): void }>()

function onClick(evt: MouseEvent) {
  if (props.loading || props.disabled) return
  emit('click', evt)
}
</script>
