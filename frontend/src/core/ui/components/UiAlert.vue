<template>
  <div :class="containerClass">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps({
  type: { type: String as () => 'info' | 'success' | 'warning' | 'error', default: 'info' },
  variant: { type: String as () => 'info' | 'success' | 'warning' | 'error', default: undefined }
})

const alertType = computed(() => props.variant || props.type)

const containerClass = computed(() => {
  const base = 'border rounded px-3 py-2 bg-white'
  const map: Record<string, string> = {
    info: 'border-gray-300',
    success: 'border-green-300',
    warning: 'border-yellow-300',
    error: 'border-red-300'
  }
  return `${base} ${map[alertType.value] || map.info}`
})
</script>
