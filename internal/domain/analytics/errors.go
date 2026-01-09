package analytics

import "errors"

// import "errors"

var (
	ErrDomainNotFound        = errors.New("domain not found")
	ErrDomainAlreadyExists        = errors.New("domain already exists")
	ErrStaleSessionsNotFound = errors.New("stale sessions not found")
	ErrLastActiveSessionNotFound = errors.New("last active session not found")
	ErrSessionsNotFound = errors.New("sessions not found")
	ErrSessionInvalid = errors.New("session invalid")
	ErrRecordEventsNotFound = errors.New("record events not found")
	ErrGuestsNotFound = errors.New("guests not found")
)
