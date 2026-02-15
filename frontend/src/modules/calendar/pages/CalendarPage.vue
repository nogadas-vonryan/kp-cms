<template>
  <div class="space-y-4">
    <div class="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
      <h1 class="text-xl sm:text-2xl font-bold text-gray-900">Calendar</h1>
      
      <!-- Mobile/Compact Status Bar -->
      <button 
        @click="showSettingsModal = true"
        class="flex items-center gap-2 px-3 py-1.5 rounded-lg text-sm font-medium transition-colors"
        :class="isLoadingOAuth ? 'bg-gray-100 text-gray-500' : (isConnected ? 'bg-green-100 text-green-800 hover:bg-green-200' : 'bg-gray-100 text-gray-700 hover:bg-gray-200')"
        :disabled="isLoadingOAuth"
      >
        <svg v-if="isLoadingOAuth" class="h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <svg v-else-if="isConnected" class="h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
        </svg>
        <svg v-else class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
        </svg>
        <span>{{ isLoadingOAuth ? 'Loading...' : (isConnected ? 'Calendar Connected' : 'Connect Calendar') }}</span>
        <Settings class="h-3.5 w-3.5 ml-1" />
      </button>
    </div>

    <CalendarEventsList @connect="navigateToSettings" />

    <!-- Settings Modal -->
    <UiModal v-model:open="showSettingsModal" title="Calendar Settings">
      <div class="space-y-6">
        <OAuthConnectionCard />

        <UiCard class="p-4 sm:p-6">
          <h3 class="text-base sm:text-lg font-semibold text-gray-900 mb-3 sm:mb-4">Quick Tips</h3>
          <ul class="space-y-2 sm:space-y-3 text-sm text-gray-600">
            <li class="flex items-start">
              <svg class="h-5 w-5 text-blue-500 mr-2 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Connect your Google Calendar to sync events automatically</span>
            </li>
            <li class="flex items-start">
              <svg class="h-5 w-5 text-blue-500 mr-2 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Create events directly from case documents</span>
            </li>
            <li class="flex items-start">
              <svg class="h-5 w-5 text-blue-500 mr-2 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span>Events are tagged for easy filtering in Google Calendar</span>
            </li>
          </ul>
        </UiCard>
      </div>
    </UiModal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { Settings } from 'lucide-vue-next';
import { useAuthStore } from '@/modules/auth/store';
import { OAuthService } from '@/modules/auth/services/oauthService';
import CalendarEventsList from '@/modules/calendar/components/CalendarEventsList.vue';
import OAuthConnectionCard from '@/modules/auth/components/OAuthConnectionCard.vue';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiModal from '@/core/ui/components/UiModal.vue';

const router = useRouter();
const authStore = useAuthStore();
const showSettingsModal = ref(false);
const isLoadingOAuth = ref(true);

const isConnected = computed(() => authStore.hasCalendarScope);

onMounted(async () => {
  try {
    const status = await OAuthService.getStatus('google');
    authStore.setOAuthConnections(status.connections);
  } catch (err) {
    console.error('Failed to fetch OAuth status:', err);
  } finally {
    isLoadingOAuth.value = false;
  }
});

function navigateToSettings() {
  router.push('/settings');
}
</script>
