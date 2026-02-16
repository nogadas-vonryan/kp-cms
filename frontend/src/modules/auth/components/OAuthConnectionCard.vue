<template>
  <UiCard class="p-6">
    <div class="space-y-4">
      <div class="flex flex-col 2xl:flex-row 2xl:items-center justify-between gap-3">
        <div class="flex items-center space-x-3">
          <div class="h-10 w-10 rounded-lg bg-white border border-gray-200 flex items-center justify-center shrink-0">
            <svg class="h-6 w-6" viewBox="0 0 24 24">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z"/>
            </svg>
          </div>
          <div>
            <h3 class="text-lg font-semibold text-gray-900">Google Calendar</h3>
            <p class="text-sm text-gray-600">Sync events with your Google Calendar</p>
          </div>
        </div>
        <div class="flex items-center shrink-0">
          <span 
            v-if="calendarStore.isLoading"
            class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800"
          >
            <svg class="animate-spin -ml-1 mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Checking...
          </span>
          <span 
            v-else-if="calendarStore.isConnected"
            class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800"
          >
            <svg class="mr-1.5 h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
            </svg>
            Connected
          </span>
          <span 
            v-else
            class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-gray-100 text-gray-800"
          >
            Not Connected
          </span>
        </div>
      </div>

      <UiAlert v-if="error" variant="error" class="mt-4">
        {{ error }}
      </UiAlert>

      <div class="border-t border-gray-200 pt-4 mt-4">
        <div class="flex flex-col 2xl:flex-row 2xl:items-center justify-between gap-3">
          <div class="text-xs text-gray-600">
            <template v-if="calendarStore.isConnected">
              <p>Your calendar is connected. You can create events from documents.</p>
              <p v-if="lastConnected" class="text-xs text-gray-500 mt-1">
                Connected since {{ formatDate(lastConnected) }}
              </p>
            </template>
            <template v-else>
              <p>Connect your Google Calendar to create events from case documents.</p>
            </template>
          </div>
          <div class="shrink-0">
            <UiButton
              v-if="calendarStore.isConnected"
              variant="danger"
              :loading="isDisconnecting"
              @click="handleDisconnect"
            >
              Disconnect
            </UiButton>
            <UiButton
              v-else
              variant="primary"
              :loading="isConnecting"
              @click="handleConnect"
            >
              Connect Google Calendar
            </UiButton>
          </div>
        </div>
      </div>
    </div>
  </UiCard>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { OAuthService } from '@/modules/auth/services/oauthService';
import { useCalendarStore } from '@/modules/calendar/store';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';

const emit = defineEmits<{
  connected: [];
  disconnected: [];
}>();

const calendarStore = useCalendarStore();

const isConnecting = ref(false);
const isDisconnecting = ref(false);
const error = ref('');

const lastConnected = computed(() => calendarStore.connection?.connected_at);

onMounted(async () => {
  await calendarStore.fetchConnectionStatus();
});

async function handleConnect() {
  try {
    isConnecting.value = true;
    error.value = '';

    const authUrl = await OAuthService.getAuthURL('google');

    const result = await OAuthService.openOAuthPopup(authUrl);

    if (result.success) {
      await calendarStore.refreshConnectionStatus();
      emit('connected');
    } else {
      error.value = result.error || 'Failed to connect Google Calendar';
    }
  } catch (err: any) {
    error.value = err.message || 'Failed to connect Google Calendar';
    console.error('Failed to connect OAuth:', err);
  } finally {
    isConnecting.value = false;
  }
}

async function handleDisconnect() {
  try {
    isDisconnecting.value = true;
    error.value = '';

    await OAuthService.disconnect('google');
    calendarStore.clearConnection();
    emit('disconnected');
  } catch (err: any) {
    error.value = 'Failed to disconnect Google Calendar';
    console.error('Failed to disconnect OAuth:', err);
  } finally {
    isDisconnecting.value = false;
  }
}

function formatDate(dateString: string): string {
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  } catch {
    return dateString;
  }
}
</script>
