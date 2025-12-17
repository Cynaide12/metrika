package analytics

import "errors"

// import "errors"

var (
	ErrDomainNotFound        = errors.New("domain not found")
	ErrStaleSessionsNotFound = errors.New("stale sessions not found")
)
