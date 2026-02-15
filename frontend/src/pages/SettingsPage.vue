<template>
  <div class="space-y-4">
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-900">Settings</h1>
    </div>

    <!-- Account Section -->
    <UiCard class="p-6">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">Account Information</h2>
      <div class="space-y-3">
        <div class="flex justify-between items-center py-2 border-b border-gray-100">
          <span class="text-gray-600">Username</span>
          <span class="font-medium text-gray-900">{{ username }}</span>
        </div>
        <div class="flex justify-between items-center py-2 border-b border-gray-100">
          <span class="text-gray-600">Role</span>
          <span class="font-medium text-gray-900 capitalize">{{ role?.replace('Role', '') }}</span>
        </div>
      </div>
    </UiCard>

    <!-- OAuth Connections Section -->
    <UiCard class="p-6">
      <h2 class="text-lg font-semibold text-gray-900 mb-4">Connected Services</h2>
      <OAuthConnectionCard />
    </UiCard>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useAuthStore } from '@/modules/auth/store';
import OAuthConnectionCard from '@/modules/auth/components/OAuthConnectionCard.vue';
import UiCard from '@/core/ui/components/UiCard.vue';

const authStore = useAuthStore();

const username = computed(() => authStore.username);
const role = computed(() => authStore.role);
</script>
