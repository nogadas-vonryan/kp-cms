<template>
  <UiCard class="p-4 sm:p-6 h-full">
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 sm:gap-4 mb-4 sm:mb-6">
      <div>
        <h3 class="text-base sm:text-lg font-semibold text-gray-900">Upcoming Events</h3>
        <p class="text-xs sm:text-sm text-gray-600">Events from your connected Google Calendar</p>
      </div>
      <UiButton 
        v-if="isConnected" 
        variant="primary" 
        size="sm"
        @click="openCreateModal"
        class="w-full sm:w-auto"
      >
        Create Event
      </UiButton>
    </div>

    <!-- Calendar Selector -->
    <div v-if="isConnected && calendars.length > 0" class="mb-4 sm:mb-6">
      <label class="block text-sm font-medium text-gray-700 mb-1">Select Calendar</label>
      <select 
        v-model="selectedCalendarId" 
        class="w-full sm:max-w-md rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-10"
        :disabled="isLoadingCalendars"
      >
        <option v-for="cal in calendars" :key="cal.id" :value="cal.id">
          {{ cal.summary }} {{ cal.primary ? '(Primary)' : '' }}
        </option>
      </select>
      <p v-if="selectedCalendar" class="mt-1 text-xs text-gray-500">
        Showing events from: {{ selectedCalendar.summary }}
      </p>
    </div>

    <!-- Loading State -->
    <div v-if="isLoading" class="flex justify-center py-6 sm:py-8">
      <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
    </div>

    <!-- Error State -->
    <UiAlert v-else-if="error" variant="error" class="mb-3 sm:mb-4">
      {{ error }}
    </UiAlert>

    <!-- Not Connected State -->
    <div v-else-if="!isConnected" class="text-center py-6 sm:py-8">
      <div class="text-gray-400 mb-3">
        <svg class="mx-auto h-10 w-10 sm:h-12 sm:w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
        </svg>
      </div>
      <p class="text-sm sm:text-base text-gray-600 mb-4 px-4">Connect your Google Calendar to view and create events</p>
      <UiButton variant="primary" @click="$emit('connect')" class="w-full sm:w-auto">
        Connect Calendar
      </UiButton>
    </div>

    <!-- Events List -->
    <div v-else-if="events.length > 0" class="space-y-2 sm:space-y-3">
      <div 
        v-for="event in events" 
        :key="event.id"
        class="flex items-start space-x-3 sm:space-x-4 p-3 sm:p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
      >
        <div class="shrink-0 w-12 sm:w-14 text-center">
          <div class="text-[10px] sm:text-xs font-medium text-gray-500 uppercase">
            {{ formatMonth(event.start_time) }}
          </div>
          <div class="text-xl sm:text-2xl font-bold text-gray-900">
            {{ formatDay(event.start_time) }}
          </div>
        </div>
        <div class="flex-1 min-w-0">
          <h4 class="text-sm font-medium text-gray-900 wrap-break-word">
            {{ event.summary }}
          </h4>
          <p v-if="event.description" class="text-xs sm:text-sm text-gray-600 mt-0.5 sm:mt-1 line-clamp-2">
            {{ event.description }}
          </p>
          <div class="flex flex-col sm:flex-row sm:items-center mt-1.5 sm:mt-2 space-y-1 sm:space-y-0 sm:space-x-4 text-xs text-gray-500">
            <span class="flex items-center">
              <svg class="mr-1 h-3 w-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span v-if="event.is_all_day" class="truncate">All day</span>
              <span v-else class="truncate">{{ formatTime(event.start_time) }} - {{ formatTime(event.end_time) }}</span>
            </span>
            <span v-if="event.location" class="flex items-center">
              <svg class="mr-1 h-3 w-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
              <span class="truncate">{{ event.location }}</span>
            </span>
          </div>
          <div v-if="event.attendees && event.attendees.length > 0" class="mt-1.5 sm:mt-2">
            <span class="text-xs text-gray-500">
              {{ event.attendees.length }} attendee(s)
            </span>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty State -->
    <div v-else class="text-center py-6 sm:py-8">
      <div class="text-gray-400 mb-3">
        <svg class="mx-auto h-10 w-10 sm:h-12 sm:w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
        </svg>
      </div>
      <p class="text-sm sm:text-base text-gray-600">No upcoming events</p>
      <p class="text-xs sm:text-sm text-gray-500 mt-1 px-4">Events created from documents will appear here</p>
    </div>
  </UiCard>

  <!-- Create Event Modal -->
  <UiModal :open="showCreateModal" title="Create Calendar Event" @close="closeCreateModal">
    <UiAlert v-if="error" variant="error" class="mb-4">
      {{ error }}
    </UiAlert>
    <div class="space-y-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Calendar</label>
        <select
          v-model="newEvent.calendar_id"
          class="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 min-h-10"
          :disabled="isLoadingCalendars"
        >
          <option v-if="isLoadingCalendars" value="">Loading calendars...</option>
          <option v-for="cal in calendars" :key="cal.id" :value="cal.id">
            {{ cal.summary }} {{ cal.primary ? '(Primary)' : '' }}
          </option>
        </select>
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Event Title</label>
        <UiInput v-model="newEvent.summary" placeholder="Enter event title" />
      </div>
      <div>
        <label class="flex items-center space-x-2 cursor-pointer">
          <input 
            v-model="newEvent.is_all_day" 
            type="checkbox" 
            class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
          />
          <span class="text-sm font-medium text-gray-700">All-day event</span>
        </label>
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Description</label>
        <UiTextarea v-model="newEvent.description" placeholder="Enter event description" :rows="3" />
      </div>
      <div :class="newEvent.is_all_day ? '' : 'grid grid-cols-1 sm:grid-cols-2 gap-3 sm:gap-4'">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Date</label>
          <UiInput v-model="newEvent.date" type="date" class="min-h-10" />
        </div>
        <div v-if="!newEvent.is_all_day">
          <label class="block text-sm font-medium text-gray-700 mb-1">Time</label>
          <UiInput v-model="newEvent.time" type="time" class="min-h-10" />
        </div>
      </div>
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-1">Location (Optional)</label>
        <UiInput v-model="newEvent.location" placeholder="Enter location" />
      </div>
    </div>
    <template #footer>
      <div class="flex flex-col-reverse sm:flex-row sm:justify-end gap-2 sm:gap-3">
        <UiButton variant="ghost" @click="closeCreateModal" class="w-full sm:w-auto">Cancel</UiButton>
        <UiButton variant="primary" :loading="isCreating" @click="handleCreate" class="w-full sm:w-auto">Create Event</UiButton>
      </div>
    </template>
  </UiModal>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useCalendarStore } from '@/modules/calendar/store';
import { CalendarService } from '@/modules/calendar/services/calendarService';
import type { Calendar } from '@/types';
import UiCard from '@/core/ui/components/UiCard.vue';
import UiButton from '@/core/ui/components/UiButton.vue';
import UiAlert from '@/core/ui/components/UiAlert.vue';
import UiInput from '@/core/ui/components/UiInput.vue';
import UiTextarea from '@/core/ui/components/UiTextarea.vue';
import UiModal from '@/core/ui/components/UiModal.vue';

const props = defineProps<{
  autoLoad?: boolean;
}>();

const emit = defineEmits<{
  (e: 'connect'): void;
}>();

const calendarStore = useCalendarStore();

const isLoading = computed(() => calendarStore.eventsLoading);
const isLoadingCalendars = computed(() => calendarStore.calendarsLoading);
const error = ref('');
const isCreating = ref(false);
const events = computed(() => calendarStore.events);
const calendars = computed(() => calendarStore.calendars);
const showCreateModal = ref(false);
const selectedCalendarId = ref<string>('');

const LOCAL_STORAGE_KEY = 'kpcms_selected_calendar_id';

const isConnected = computed(() => calendarStore.isConnected);

const selectedCalendar = computed(() => {
  return calendars.value.find(cal => cal.id === selectedCalendarId.value);
});

function loadSelectedCalendarFromStorage(): string | null {
  try {
    return localStorage.getItem(LOCAL_STORAGE_KEY);
  } catch {
    return null;
  }
}

function saveSelectedCalendarToStorage(calendarId: string) {
  try {
    localStorage.setItem(LOCAL_STORAGE_KEY, calendarId);
  } catch (err) {
    console.warn('Failed to save calendar selection to localStorage:', err);
  }
}

function findDefaultCalendar(): Calendar | undefined {
  // First try to find "Katarungang Pambarangay" calendar
  const kpCalendar = calendars.value.find(cal => 
    cal.summary.toLowerCase().includes('katarungang pambarangay')
  );
  if (kpCalendar) return kpCalendar;
  
  // Fall back to primary calendar
  const primaryCalendar = calendars.value.find(cal => cal.primary);
  if (primaryCalendar) return primaryCalendar;
  
  // Last resort: first available calendar
  return calendars.value[0];
}

const newEvent = ref({
  summary: '',
  description: '',
  date: new Date().toISOString().split('T')[0],
  time: '10:00',
  location: '',
  calendar_id: '',
  is_all_day: false,
});

onMounted(async () => {
  if (isConnected.value) {
    await calendarStore.fetchCalendars();
    if (props.autoLoad !== false) {
      await calendarStore.fetchEvents(selectedCalendarId.value || undefined);
    }
    initializeSelectedCalendar();
  }
});

watch(() => isConnected.value, async (connected) => {
  if (connected) {
    await calendarStore.fetchCalendars();
    if (props.autoLoad !== false) {
      await calendarStore.fetchEvents(selectedCalendarId.value || undefined);
    }
    initializeSelectedCalendar();
  }
});

function initializeSelectedCalendar() {
  const savedCalendarId = loadSelectedCalendarFromStorage();
  
  if (savedCalendarId && calendars.value.some(cal => cal.id === savedCalendarId)) {
    selectedCalendarId.value = savedCalendarId;
  } else {
    const defaultCalendar = findDefaultCalendar();
    if (defaultCalendar) {
      selectedCalendarId.value = defaultCalendar.id;
      saveSelectedCalendarToStorage(defaultCalendar.id);
    }
  }
  
  // Sync with store for refresh functionality
  if (selectedCalendarId.value) {
    calendarStore.selectedCalendarId = selectedCalendarId.value;
  }
}

watch(selectedCalendarId, (newCalendarId) => {
  if (newCalendarId) {
    saveSelectedCalendarToStorage(newCalendarId);
    calendarStore.fetchEvents(newCalendarId || undefined);
  }
});

async function openCreateModal() {
  error.value = '';
  showCreateModal.value = true;
  await calendarStore.fetchCalendars();
  const defaultCalendar = findDefaultCalendar();
  if (defaultCalendar) {
    newEvent.value.calendar_id = defaultCalendar.id;
  }
}

function closeCreateModal() {
  error.value = '';
  showCreateModal.value = false;
}

async function handleCreate() {
  // Clear any previous errors at the start
  error.value = '';
  
  if (!newEvent.value.summary) {
    error.value = 'Event title is required';
    return;
  }

  try {
    isCreating.value = true;

    let startDate: Date;
    let endDate: Date;

    if (newEvent.value.is_all_day) {
      // For all-day events, use date without time component
      const dateStr = (newEvent.value.date || new Date().toISOString().split('T')[0]) as string;
      startDate = new Date(dateStr);
      // End date is the next day for all-day events
      endDate = new Date(startDate);
      endDate.setDate(endDate.getDate() + 1);
    } else {
      startDate = new Date(`${newEvent.value.date}T${newEvent.value.time}`);
      endDate = new Date(startDate.getTime() + 60 * 60 * 1000); // 1 hour default duration
    }

    await CalendarService.createEvent({
      summary: newEvent.value.summary,
      description: newEvent.value.description,
      location: newEvent.value.location || undefined,
      start_time: startDate.toISOString(),
      end_time: endDate.toISOString(),
      calendar_id: newEvent.value.calendar_id || undefined,
      is_all_day: newEvent.value.is_all_day,
    });

    // Reset form and close modal
    // Preserve the currently selected calendar_id if available
    const currentCalendarId = newEvent.value.calendar_id || calendars.value.find(cal => cal.primary)?.id || '';
    newEvent.value = {
      summary: '',
      description: '',
      date: new Date().toISOString().split('T')[0],
      time: '10:00',
      location: '',
      calendar_id: currentCalendarId,
      is_all_day: false,
    };
    showCreateModal.value = false;

    // Reload events
    await calendarStore.fetchEvents(newEvent.value.calendar_id || undefined);
  } catch (err: any) {
    error.value = err.response?.data?.error || 'Failed to create event';
    console.error('Failed to create calendar event:', err);
  } finally {
    isCreating.value = false;
  }
}

function formatMonth(dateString: string): string {
  try {
    const date = new Date(dateString);
    return date.toLocaleDateString('en-US', { month: 'short' });
  } catch {
    return '';
  }
}

function formatDay(dateString: string): string {
  try {
    const date = new Date(dateString);
    return date.getDate().toString();
  } catch {
    return '';
  }
}

function formatTime(dateString: string): string {
  try {
    const date = new Date(dateString);
    return date.toLocaleTimeString('en-US', { 
      hour: 'numeric', 
      minute: '2-digit',
      hour12: true 
    });
  } catch {
    return '';
  }
}
</script>
