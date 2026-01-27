<template>
  <div class="min-h-[50vh] flex items-center justify-center">
    <div class="text-center space-y-4">
      <h1 class="text-6xl font-bold text-gray-900">404</h1>
      <p class="text-xl text-gray-600">Page not found</p>
      <UiButton @click="$router.push('/documents')">
        Go to Documents
      </UiButton>
    </div>
  </div>
</template>

<script setup lang="ts">
import UiButton from '@/core/ui/components/UiButton.vue';
</script>
