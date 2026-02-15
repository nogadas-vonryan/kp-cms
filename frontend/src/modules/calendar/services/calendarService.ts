import api from '@/core/api/client';
import type { 
  CalendarEvent, 
  CreateCalendarEventRequest, 
  ListCalendarEventsResponse,
  ListCalendarsResponse,
  Calendar
} from '@/types';

// Default event duration in milliseconds (1 hour)
const DEFAULT_EVENT_DURATION_MS = 60 * 60 * 1000;

export interface ListEventsOptions {
  maxResults?: number;
  startTime?: Date;
  endTime?: Date;
  showDeleted?: boolean;
  calendarId?: string;
}

export const CalendarService = {
  /**
   * Get calendar connection status
   * @returns Connection status and details
   */
  async getStatus(): Promise<{ is_connected: boolean; error?: string }> {
    try {
      const response = await api.get('/api/calendar/status');
      return response.data;
    } catch (err: any) {
      return { 
        is_connected: false, 
        error: err.response?.data?.error || 'Failed to check calendar status'
      };
    }
  },

  /**
   * List calendar events
   * @param options - Query options for listing events
   * @returns List of events and pagination token
   */
  async listEvents(options: ListEventsOptions = {}): Promise<ListCalendarEventsResponse> {
    const params = new URLSearchParams();
    
    if (options.maxResults) {
      params.append('max_results', options.maxResults.toString());
    }
    if (options.startTime) {
      params.append('start_time', options.startTime.toISOString());
    }
    if (options.endTime) {
      params.append('end_time', options.endTime.toISOString());
    }
    if (options.showDeleted) {
      params.append('showDeleted', 'true');
    }
    if (options.calendarId) {
      params.append('calendar_id', options.calendarId);
    }

    const response = await api.get<ListCalendarEventsResponse>(`/api/calendar/events?${params.toString()}`);
    return response.data;
  },

  /**
   * List available calendars
   * @returns List of calendars
   */
  async listCalendars(): Promise<Calendar[]> {
    const response = await api.get<ListCalendarsResponse>('/api/calendar/calendars');
    return response.data.calendars;
  },

  /**
   * Create a new calendar event
   * @param event - Event details
   * @returns Created event
   */
  async createEvent(event: CreateCalendarEventRequest): Promise<CalendarEvent> {
    const response = await api.post<CalendarEvent>('/api/calendar/events', event);
    return response.data;
  },

  /**
   * Create a calendar event from a case document
   * @param documentUuid - Document UUID
   * @param scheduledDate - Date for the hearing/event
   * @param additionalDetails - Optional additional details
   * @returns Created event
   */
  async createEventFromDocument(
    documentUuid: string,
    scheduledDate: Date,
    additionalDetails?: {
      title?: string;
      description?: string;
      location?: string;
      attendees?: string[];
    }
  ): Promise<CalendarEvent> {
    const event: CreateCalendarEventRequest = {
      summary: additionalDetails?.title || 'Case Hearing',
      description: additionalDetails?.description || `Case document: ${documentUuid}`,
      location: additionalDetails?.location,
      start_time: scheduledDate.toISOString(),
      end_time: new Date(scheduledDate.getTime() + DEFAULT_EVENT_DURATION_MS).toISOString(),
      attendees: additionalDetails?.attendees,
    };

    return this.createEvent(event);
  }
};
