<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/20" @click="handleOverlayClick" />
      
      <div :class="[
        'relative border border-gray-300 rounded bg-white w-full flex flex-col max-h-[90vh]',
        modalSizeClass
      ]">
        
        <div v-if="title" class="p-4 border-b border-gray-200 font-medium text-gray-900 flex justify-between items-center">
          <span>{{ title }}</span>
          <button v-if="!preventClose" @click="close" class="text-gray-400 hover:text-gray-600">×</button>
        </div>

        <div class="flex-1 overflow-y-auto p-4 custom-scrollbar">
          <slot />
        </div>

        <div class="p-4 border-t border-gray-200 flex justify-end gap-2 bg-gray-50/50">
          <slot name="footer">
            <button 
              class="border border-gray-300 rounded px-3 py-2 bg-white disabled:opacity-50 disabled:cursor-not-allowed" 
              :disabled="preventClose"
              @click="close"
            >
              Close
            </button>
          </slot>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' },
  preventClose: { type: Boolean, default: false },
  size: { type: String, default: 'md' } // sm, md, lg, xl
})

const emit = defineEmits<{ (e: 'update:open', value: boolean): void; (e: 'close'): void }>()

const modalSizeClass = computed(() => {
  const sizes = {
    sm: 'max-w-sm',
    md: 'max-w-lg',
    lg: 'max-w-2xl',
    xl: 'max-w-4xl'
  };
  return sizes[props.size as keyof typeof sizes] || sizes.md;
});

function handleOverlayClick() {
  if (!props.preventClose) {
    close()
  }
}

function close() {
  if (props.preventClose) return
  emit('update:open', false)
  emit('close')
}
</script>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 5px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #d1d5db; /* gray-300 */
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #9ca3af; /* gray-400 */
}
</style>