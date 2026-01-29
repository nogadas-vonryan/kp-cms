<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import AppShell from './layouts/AppShell.vue'

const route = useRoute();
const authStore = useAuthStore();

const isPublicRoute = computed(() => route.meta.public);
const showAppShell = computed(() => !isPublicRoute.value && authStore.isAuthenticated);
</script>

<template>
  <AppShell v-if="showAppShell" />
  <div v-else class="min-h-screen bg-gray-50">
    <RouterView />
  </div>
</template>