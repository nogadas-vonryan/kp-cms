<template>
  <button
    :type="type"
    :disabled="disabled || loading"
    class="inline-flex items-center justify-center whitespace-nowrap border rounded px-3 py-2 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
    :class="[
      block ? 'w-full' : 'w-auto',
      variantClasses[variant] // This injects the colors based on the prop
    ]"
    @click="onClick"
  >
    <span v-if="loading" class="mr-2 h-3 w-3 rounded-full border-2 border-t-transparent animate-spin" 
      :class="variant === 'primary' || variant === 'danger' ? 'border-white/30 border-t-white' : 'border-gray-300 border-t-blue-600'">
    </span>
    <slot />
  </button>
</template>

<script setup lang="ts">

const props = defineProps({
  type: { 
    type: String as () => 'button' | 'submit' | 'reset', 
    default: 'button' 
  },
  variant: { 
    type: String as () => 'default' | 'primary' | 'secondary' | 'danger' | 'ghost', 
    default: 'default' 
  },
  disabled: { type: Boolean, default: false },
  loading: { type: Boolean, default: false },
  block: { type: Boolean, default: false }
})

const variantClasses = {
  default: 'border-gray-300 bg-white text-gray-900 hover:bg-gray-50',
  primary: 'border-blue-600 bg-blue-600 text-white hover:bg-blue-700',
  secondary: 'border-gray-800 bg-gray-800 text-white hover:bg-gray-900', // Or a Slate/Gray color
  danger: 'border-red-600 bg-red-500 text-white hover:bg-red-600',
  ghost: 'border-transparent bg-transparent text-gray-600 hover:bg-gray-100'
}

const emit = defineEmits<{ (e: 'click', evt: MouseEvent): void }>()

function onClick(evt: MouseEvent) {
  if (props.loading || props.disabled) return
  emit('click', evt)
}
</script>