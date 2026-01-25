<script setup lang="ts">
import { RouterView, useRouter } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import { AuthService } from '@/modules/auth/services/authService';
import { computed } from 'vue';

const router = useRouter();
const authStore = useAuthStore();

const isAuthenticated = computed(() => authStore.isAuthenticated);
const isAdmin = computed(() => authStore.role === 'RoleAdmin');
const username = computed(() => authStore.user?.id || '');

async function logout() {
  await AuthService.logout();
  router.push('/login');
}
</script>

<template>
  <div class="min-h-screen bg-gray-50">
    <!-- Navigation Header -->
    <nav v-if="isAuthenticated" class="bg-white border-b border-gray-200">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="flex justify-between h-16">
          <div class="flex">
            <div class="shrink-0 flex items-center">
              <span class="text-xl font-bold text-gray-900">Archivist</span>
            </div>
            <div class="hidden sm:ml-6 sm:flex sm:space-x-8">
              <router-link
                to="/documents"
                class="inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                active-class="border-blue-500 text-gray-900"
                inactive-class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              >
                Documents
              </router-link>
              <router-link
                v-if="isAdmin"
                to="/admin"
                class="inline-flex items-center px-1 pt-1 border-b-2 text-sm font-medium"
                active-class="border-blue-500 text-gray-900"
                inactive-class="border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700"
              >
                Admin
              </router-link>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <span class="text-sm text-gray-700">{{ username }}</span>
            <button
              @click="logout"
              class="text-sm text-gray-600 hover:text-gray-900"
            >
              Logout
            </button>
          </div>
        </div>
      </div>
    </nav>

    <!-- Main Content -->
    <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <RouterView />
    </main>
  </div>
</template>
