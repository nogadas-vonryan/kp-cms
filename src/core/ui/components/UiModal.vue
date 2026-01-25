<template>
  <Teleport to="body">
    <div v-if="open" class="fixed inset-0 z-50">
      <div class="absolute inset-0 bg-black/20" @click="close" />
      <div class="absolute inset-0 flex items-center justify-center p-4">
        <div class="border border-gray-300 rounded bg-white w-full max-w-lg p-4">
          <div v-if="title" class="mb-2 font-medium text-gray-900">{{ title }}</div>
          <slot />
          <div class="mt-4 flex justify-end gap-2">
            <slot name="footer">
              <button class="border border-gray-300 rounded px-3 py-2 bg-white" @click="close">Close</button>
            </slot>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
const props = defineProps({
  open: { type: Boolean, default: false },
  title: { type: String, default: '' }
})

const emit = defineEmits<{ (e: 'update:open', value: boolean): void; (e: 'close'): void }>()

function close() {
  emit('update:open', false)
  emit('close')
}
</script>
