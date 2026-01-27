<template>
  <div 
    v-if="modelValue" 
    :class="[
      'p-3 border rounded transition-all',
      config.containerClass
    ]"
  >
    <div class="flex items-start justify-between">
      <div class="flex-1">
        <div class="flex items-center gap-2">
          <span :class="['inline-flex items-center px-2 py-1 rounded text-xs font-medium uppercase tracking-tight', config.badgeClass]">
            {{ label || type }}
          </span>
          <slot name="title">
            <span class="text-sm font-medium text-gray-900">{{ title }}</span>
          </slot>
        </div>
        <div v-if="$slots.default" class="mt-1">
          <slot />
        </div>
      </div>
      <button 
        v-if="dismissible"
        @click="$emit('update:modelValue', null)" 
        class="text-gray-400 hover:text-gray-600 text-[10px] font-mono uppercase pt-1"
      >
        [ Dismiss ]
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

// Define a proper interface for your message object if applicable
interface NoticeProps {
  modelValue?: any;
  type?: 'success' | 'error' | 'warning'; // Made optional because we provide a default
  label?: string;
  title?: string;
  dismissible?: boolean;
}

const props = withDefaults(defineProps<NoticeProps>(), {
  type: 'success', // Default fallback handled here
  dismissible: true
});

const emit = defineEmits(['update:modelValue']);

const config = computed(() => {
  const styles = {
    success: { containerClass: 'bg-green-50 border-green-300', badgeClass: 'bg-green-200 text-green-800' },
    error: { containerClass: 'bg-red-50 border-red-300', badgeClass: 'bg-red-200 text-red-800' },
    warning: { containerClass: 'bg-yellow-50 border-yellow-300', badgeClass: 'bg-yellow-200 text-yellow-800' }
  };
  // Fallback to success if type is undefined/invalid
  return styles[props.type] || styles.success;
});
</script>