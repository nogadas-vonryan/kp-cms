<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
    <UiCard class="w-full max-w-md">
      <div class="space-y-6">
        <div class="text-center">
          <h1 class="text-2xl font-bold text-gray-900">Archivist</h1>
          <p class="mt-2 text-sm text-gray-600">Sign in to your account</p>
        </div>

        <form @submit.prevent="handleSubmit" class="space-y-4">
          <div>
            <label for="username" class="block text-sm font-medium text-gray-700 mb-1">
              Username
            </label>
            <UiInput
              id="username"
              v-model="username"
              type="text"
              placeholder="Enter your username"
              required
              autocomplete="username"
              :disabled="loading"
            />
          </div>

          <div>
            <label for="password" class="block text-sm font-medium text-gray-700 mb-1">
              Password
            </label>
            <UiInput
              id="password"
              v-model="password"
              type="password"
              placeholder="••••••••"
              required
              autocomplete="current-password"
              :disabled="loading"
            />
          </div>

          <UiAlert v-if="error" variant="error">
            {{ error }}
          </UiAlert>

          <UiButton
            type="submit"
            :loading="loading"
            :disabled="loading"
            block
          >
            Sign In
          </UiButton>
        </form>
      </div>
    </UiCard>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { AuthService } from '@/modules/auth/services/authService';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';

const router = useRouter();
const route = useRoute();

const username = ref('');
const password = ref('');
const loading = ref(false);
const error = ref('');

async function handleSubmit() {
  error.value = '';
  loading.value = true;

  try {
    await AuthService.login({ username: username.value, password: password.value });
    const redirect = (route.query.redirect as string) || '/documents';
    router.push(redirect);
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Login failed. Please try again.';
  } finally {
    loading.value = false;
  }
}
</script>
