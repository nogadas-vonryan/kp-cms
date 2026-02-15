package calendar

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"kpcms/server/core/oauth"

	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

// Service provides calendar functionality using the OAuth service
type Service struct {
	oauthService *oauth.Service
	appID        string
}

// New creates a new calendar service
func New(oauthService *oauth.Service, appID string) *Service {
	if appID == "" {
		appID = "kpcms-calendar"
	}
	return &Service{
		oauthService: oauthService,
		appID:        appID,
	}
}

// IsConnected checks if a user has connected their Google Calendar
func (s *Service) IsConnected(username string) bool {
	return s.oauthService.IsConnected(username, oauth.ProviderGoogle) &&
		s.oauthService.HasScope(username, oauth.ProviderGoogle, oauth.ScopeCalendar)
}

// GetConnectionStatus returns the calendar connection status for a user
func (s *Service) GetConnectionStatus(username string) (*oauth.Connection, error) {
	connections, err := s.oauthService.GetConnections(username)
	if err != nil {
		return nil, err
	}

	for _, conn := range connections {
		if conn.Provider == string(oauth.ProviderGoogle) {
			// Check if calendar scope is granted
			for _, scope := range conn.Scopes {
				if scope == oauth.ScopeCalendar {
					return conn, nil
				}
			}
		}
	}

	return &oauth.Connection{
		Username:    username,
		Provider:    string(oauth.ProviderGoogle),
		IsConnected: false,
		Scopes:      []string{},
	}, nil
}

// CreateEvent creates a new calendar event
func (s *Service) CreateEvent(ctx context.Context, username string, req *CreateEventRequest) (*Event, error) {
	slog.Info("Creating calendar event",
		"username", username,
		"summary", req.Summary,
	)

	// Get HTTP client from OAuth service
	client, err := s.oauthService.GetHTTPClient(ctx, username, oauth.ProviderGoogle)
	if err != nil {
		slog.Error("Failed to get HTTP client",
			"username", username,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get calendar client: %w", err)
	}

	// Create calendar service
	calService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Build the Google Calendar event
	gcalEvent := &calendar.Event{
		Summary:     req.Summary,
		Description: req.Description,
		Location:    req.Location,
		ExtendedProperties: &calendar.EventExtendedProperties{
			Private: map[string]string{
				"app_id":   s.appID,
				"username": username,
			},
		},
	}

	// Set start time
	if req.IsAllDay {
		gcalEvent.Start = &calendar.EventDateTime{
			Date: req.StartTime.Format("2006-01-02"),
		}
		gcalEvent.End = &calendar.EventDateTime{
			Date: req.EndTime.Format("2006-01-02"),
		}
	} else {
		gcalEvent.Start = &calendar.EventDateTime{
			DateTime: req.StartTime.Format(time.RFC3339),
		}
		gcalEvent.End = &calendar.EventDateTime{
			DateTime: req.EndTime.Format(time.RFC3339),
		}
	}

	// Add attendees if provided
	if len(req.Attendees) > 0 {
		gcalAttendees := make([]*calendar.EventAttendee, len(req.Attendees))
		for i, email := range req.Attendees {
			gcalAttendees[i] = &calendar.EventAttendee{Email: email}
		}
		gcalEvent.Attendees = gcalAttendees
	}

	// Insert the event
	createdEvent, err := calService.Events.Insert("primary", gcalEvent).Do()
	if err != nil {
		slog.Error("Failed to create calendar event",
			"username", username,
			"error", err,
		)
		return nil, fmt.Errorf("failed to create event: %w", err)
	}

	slog.Info("Event created successfully",
		"username", username,
		"event_id", createdEvent.Id,
	)

	return convertGoogleEvent(createdEvent), nil
}

// ListEvents lists calendar events
func (s *Service) ListEvents(ctx context.Context, username string, opts ListOptions) (*ListEventsResponse, error) {
	slog.Info("Listing calendar events",
		"username", username,
	)

	// Get HTTP client from OAuth service
	client, err := s.oauthService.GetHTTPClient(ctx, username, oauth.ProviderGoogle)
	if err != nil {
		slog.Error("Failed to get HTTP client",
			"username", username,
			"error", err,
		)
		return nil, fmt.Errorf("failed to get calendar client: %w", err)
	}

	// Create calendar service
	calService, err := calendar.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar service: %w", err)
	}

	// Build the query
	query := fmt.Sprintf("privateExtendedProperty.app_id=%s", s.appID)

	call := calService.Events.List("primary").
		PrivateExtendedProperty(query).
		SingleEvents(true).
		OrderBy("startTime")

	if opts.MaxResults > 0 {
		call = call.MaxResults(opts.MaxResults)
	} else {
		call = call.MaxResults(50)
	}

	if !opts.StartTime.IsZero() {
		call = call.TimeMin(opts.StartTime.Format(time.RFC3339))
	}

	if !opts.EndTime.IsZero() {
		call = call.TimeMax(opts.EndTime.Format(time.RFC3339))
	}

	if opts.ShowDeleted {
		call = call.ShowDeleted(true)
	}

	// Execute the query
	events, err := call.Do()
	if err != nil {
		slog.Error("Failed to list calendar events",
			"username", username,
			"error", err,
		)
		return nil, fmt.Errorf("failed to list events: %w", err)
	}

	slog.Info("Events listed successfully",
		"username", username,
		"count", len(events.Items),
	)

	return &ListEventsResponse{
		Events:        convertGoogleEvents(events.Items),
		NextPageToken: events.NextPageToken,
	}, nil
}

// convertGoogleEvent converts a Google Calendar event to our Event type
func convertGoogleEvent(gcalEvent *calendar.Event) *Event {
	event := &Event{
		ID:          gcalEvent.Id,
		Summary:     gcalEvent.Summary,
		Description: gcalEvent.Description,
		Location:    gcalEvent.Location,
	}

	// Parse start time
	if gcalEvent.Start != nil {
		if gcalEvent.Start.DateTime != "" {
			t, _ := time.Parse(time.RFC3339, gcalEvent.Start.DateTime)
			event.StartTime = t
		} else if gcalEvent.Start.Date != "" {
			t, _ := time.Parse("2006-01-02", gcalEvent.Start.Date)
			event.StartTime = t
			event.IsAllDay = true
		}
	}

	// Parse end time
	if gcalEvent.End != nil {
		if gcalEvent.End.DateTime != "" {
			t, _ := time.Parse(time.RFC3339, gcalEvent.End.DateTime)
			event.EndTime = t
		} else if gcalEvent.End.Date != "" {
			t, _ := time.Parse("2006-01-02", gcalEvent.End.Date)
			event.EndTime = t
		}
	}

	// Parse attendees
	if len(gcalEvent.Attendees) > 0 {
		event.Attendees = make([]string, 0, len(gcalEvent.Attendees))
		for _, attendee := range gcalEvent.Attendees {
			if attendee.Email != "" {
				event.Attendees = append(event.Attendees, attendee.Email)
			}
		}
	}

	// Parse extended properties for app metadata
	if gcalEvent.ExtendedProperties != nil && gcalEvent.ExtendedProperties.Private != nil {
		event.CreatedBy = gcalEvent.ExtendedProperties.Private["username"]
	}

	return event
}

// convertGoogleEvents converts a slice of Google Calendar events
func convertGoogleEvents(gcalEvents []*calendar.Event) []*Event {
	events := make([]*Event, 0, len(gcalEvents))
	for _, gcalEvent := range gcalEvents {
		if gcalEvent != nil {
			events = append(events, convertGoogleEvent(gcalEvent))
		}
	}
	return events
}
