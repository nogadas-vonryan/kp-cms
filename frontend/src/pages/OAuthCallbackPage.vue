<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-50 px-4">
    <UiCard class="w-full max-w-md p-8 text-center">
      <div v-if="status === 'loading'" class="space-y-4">
        <div class="flex justify-center">
          <div class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
        </div>
        <h2 class="text-xl font-semibold text-gray-900">Connecting to Google...</h2>
        <p class="text-gray-600">Please wait while we complete the authorization.</p>
      </div>

      <div v-else-if="status === 'success'" class="space-y-4">
        <div class="flex justify-center">
          <div class="rounded-full h-16 w-16 bg-green-100 flex items-center justify-center">
            <svg class="h-8 w-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
            </svg>
          </div>
        </div>
        <h2 class="text-xl font-semibold text-gray-900">Connected Successfully!</h2>
        <p class="text-gray-600">Your Google Calendar has been connected to KPCMS.</p>
        <p class="text-sm text-gray-500">This window will close automatically...</p>
      </div>

      <div v-else-if="status === 'error'" class="space-y-4">
        <div class="flex justify-center">
          <div class="rounded-full h-16 w-16 bg-red-100 flex items-center justify-center">
            <svg class="h-8 w-8 text-red-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </div>
        </div>
        <h2 class="text-xl font-semibold text-gray-900">Connection Failed</h2>
        <p class="text-gray-600">{{ errorMessage }}</p>
        <UiButton @click="closeWindow" variant="secondary">
          Close Window
        </UiButton>
      </div>
    </UiCard>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import api from '@/core/api/client';

const route = useRoute();
const status = ref<'loading' | 'success' | 'error'>('loading');
const errorMessage = ref('');

onMounted(async () => {
  const state = route.query.state as string;
  const code = route.query.code as string;
  const error = route.query.error as string;

  if (error) {
    status.value = 'error';
    errorMessage.value = `Authorization error: ${error}`;
    await notifyParent(false, error);
    return;
  }

  if (!state || !code) {
    status.value = 'error';
    errorMessage.value = 'Missing authorization parameters. Please try again.';
    await notifyParent(false, 'Missing parameters');
    return;
  }

  try {
    // Call the backend callback endpoint
    await api.get('/oauth/callback', {
      params: { state, code }
    });

    status.value = 'success';
    await notifyParent(true);

    setTimeout(() => {
      closeWindow();
    }, 1500);
  } catch (err: any) {
    status.value = 'error';
    errorMessage.value = err.response?.data?.error || 'Failed to complete authorization. Please try again.';
    await notifyParent(false, errorMessage.value);
  }
});

function notifyParent(success: boolean, error?: string): Promise<void> {
  return new Promise((resolve) => {
    const channel = new BroadcastChannel('oauth-callback');
    let acknowledged = false;

    channel.onmessage = (event) => {
      if (event.data?.type === 'oauth-ack') {
        acknowledged = true;
        channel.close();
        resolve();
      }
    };

    channel.postMessage({
      type: 'oauth-callback',
      success,
      error
    });

    setTimeout(() => {
      if (!acknowledged) {
        channel.close();
        resolve();
      }
    }, 1000);
  });
}

function closeWindow() {
  if (window.opener) {
    window.close();
  }
}
</script>
