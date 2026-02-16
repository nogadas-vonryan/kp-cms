import { defineStore } from 'pinia';
import type { OAuthConnection, Calendar, CalendarEvent } from '@/types';
import { OAuthService } from '@/modules/auth/services/oauthService';
import { CalendarService } from '@/modules/calendar/services/calendarService';

interface CalendarState {
  connection: OAuthConnection | null;
  isLoading: boolean;
  lastFetched: number | null;
  calendars: Calendar[];
  events: CalendarEvent[];
  calendarsLoading: boolean;
  eventsLoading: boolean;
  calendarsLastFetched: number | null;
  eventsLastFetched: number | null;
  selectedCalendarId: string | null;
}

const CACHE_DURATION = 5 * 60 * 1000;

export const useCalendarStore = defineStore('calendar', {
  state: (): CalendarState => ({
    connection: null,
    isLoading: false,
    lastFetched: null,
    calendars: [],
    events: [],
    calendarsLoading: false,
    eventsLoading: false,
    calendarsLastFetched: null,
    eventsLastFetched: null,
    selectedCalendarId: null,
  }),
  getters: {
    isConnected: (state): boolean => {
      if (!state.connection || !state.connection.is_connected) return false;
      return state.connection.scopes?.some(scope => scope.includes('calendar')) ?? false;
    },
    hasConnectionCache: (state): boolean => {
      if (!state.lastFetched) return false;
      return Date.now() - state.lastFetched < CACHE_DURATION;
    },
    hasCalendarsCache: (state): boolean => {
      if (!state.calendarsLastFetched) return false;
      return Date.now() - state.calendarsLastFetched < CACHE_DURATION;
    },
    hasEventsCache: (state): boolean => {
      if (!state.eventsLastFetched) return false;
      return Date.now() - state.eventsLastFetched < CACHE_DURATION;
    },
  },
  actions: {
    async fetchConnectionStatus(force = false) {
      if (!force && this.hasConnectionCache) return;

      this.isLoading = true;
      try {
        const status = await OAuthService.getStatus('google');
        const conn = status.connections.find(c => c.provider === 'google');
        this.connection = conn ?? null;
        this.lastFetched = Date.now();
      } finally {
        this.isLoading = false;
      }
    },
    async refreshConnectionStatus() {
      await this.fetchConnectionStatus(true);
    },
    async fetchCalendars(force = false) {
      if (!force && this.hasCalendarsCache) return;

      this.calendarsLoading = true;
      try {
        const calendars = await CalendarService.listCalendars();
        this.calendars = calendars;
        this.calendarsLastFetched = Date.now();
      } finally {
        this.calendarsLoading = false;
      }
    },
    async fetchEvents(calendarId?: string, force = false) {
      if (!force && this.hasEventsCache && this.selectedCalendarId === calendarId) return;

      this.eventsLoading = true;
      try {
        const response = await CalendarService.listEvents({
          maxResults: 10,
          startTime: new Date(),
          calendarId,
        });
        this.events = response.events;
        this.eventsLastFetched = Date.now();
        if (calendarId) {
          this.selectedCalendarId = calendarId;
        }
      } finally {
        this.eventsLoading = false;
      }
    },
    async refreshAll() {
      await Promise.all([
        this.refreshConnectionStatus(),
        this.fetchCalendars(true),
        this.fetchEvents(this.selectedCalendarId || undefined, true),
      ]);
    },
    setConnection(connection: OAuthConnection | null) {
      this.connection = connection;
      this.lastFetched = Date.now();
    },
    clearConnection() {
      this.connection = null;
      this.lastFetched = null;
      this.calendars = [];
      this.events = [];
      this.calendarsLastFetched = null;
      this.eventsLastFetched = null;
      this.selectedCalendarId = null;
    },
  },
});
