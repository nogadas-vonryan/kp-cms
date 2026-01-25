<template>
  <div :class="containerClass">
    <slot />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps({
  padded: { type: Boolean, default: true }
})

const containerClass = computed(() => {
  return [
    'border border-gray-300 rounded bg-white',
    props.padded ? 'p-4' : 'p-0'
  ].join(' ')
})
</script>
