<template>
  <div class="min-h-screen bg-gray-50 py-8">
    <div class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
      <div class="mb-8">
        <h1 class="text-3xl font-bold text-gray-900">Settings</h1>
        <p class="mt-2 text-gray-600">Manage your account and connected services</p>
      </div>

      <div class="space-y-6">
        <!-- Account Section -->
        <UiCard class="p-6">
          <h2 class="text-xl font-semibold text-gray-900 mb-4">Account Information</h2>
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
        <div>
          <h2 class="text-xl font-semibold text-gray-900 mb-4">Connected Services</h2>
          <OAuthConnectionCard />
        </div>

        <!-- Danger Zone -->
        <UiCard class="p-6 border-red-200">
          <h2 class="text-xl font-semibold text-red-600 mb-4">Danger Zone</h2>
          <div class="flex items-center justify-between">
            <div>
              <p class="font-medium text-gray-900">Log out</p>
              <p class="text-sm text-gray-600">Sign out of your account</p>
            </div>
            <UiButton variant="danger" @click="handleLogout">
              Log Out
            </UiButton>
          </div>
        </UiCard>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import { useRouter } from 'vue-router';
import { useAuthStore } from '@/modules/auth/store';
import { AuthService } from '@/modules/auth/services/authService';
import OAuthConnectionCard from '@/modules/auth/components/OAuthConnectionCard.vue';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiButton from '@/core/ui/components/UiButton.vue';

const router = useRouter();
const authStore = useAuthStore();

const username = computed(() => authStore.username);
const role = computed(() => authStore.role);

async function handleLogout() {
  try {
    await AuthService.logout();
    router.push('/login');
  } catch (error) {
    console.error('Logout failed:', error);
  }
}
</script>
