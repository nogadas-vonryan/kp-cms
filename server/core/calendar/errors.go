package calendar

import "errors"

var (
	// ErrNotConnected is returned when a user has not connected their calendar
	ErrNotConnected = errors.New("calendar: Google Calendar not connected")

	// ErrEventNotFound is returned when an event is not found
	ErrEventNotFound = errors.New("calendar: event not found")
)
