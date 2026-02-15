package calendar

import (
	"time"
)

// Event represents a calendar event
type Event struct {
	ID          string    `json:"id"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Attendees   []string  `json:"attendees"`
	IsAllDay    bool      `json:"is_all_day"`
	CreatedBy   string    `json:"created_by"`
	CalendarID  string    `json:"calendar_id"`
}

// ListOptions contains options for listing events
type ListOptions struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	CalendarID  string    `json:"calendar_id"`
	MaxResults  int64     `json:"max_results"`
	ShowDeleted bool      `json:"show_deleted"`
}

// Connection represents a user's calendar connection
type Connection struct {
	Username    string    `json:"username"`
	CalendarID  string    `json:"calendar_id"`
	IsConnected bool      `json:"is_connected"`
	IsPrimary   bool      `json:"is_primary"`
	ConnectedAt time.Time `json:"connected_at"`
}

// CreateEventRequest represents a request to create an event
type CreateEventRequest struct {
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Location    string    `json:"location"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Attendees   []string  `json:"attendees"`
	IsAllDay    bool      `json:"is_all_day"`
	CalendarID  string    `json:"calendar_id"`
}

// ListEventsResponse represents the response for listing events
type ListEventsResponse struct {
	Events        []*Event `json:"events"`
	NextPageToken string   `json:"next_page_token,omitempty"`
}

// CalendarConfig represents the calendar module configuration
type CalendarConfig struct {
	CredentialsPath string
	CallbackURL     string
	AppID           string
}
