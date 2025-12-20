package analytics

import "errors"

// import "errors"

var (
	ErrDomainNotFound        = errors.New("domain not found")
	ErrDomainAlreadyExists        = errors.New("domain already exists")
	ErrStaleSessionsNotFound = errors.New("stale sessions not found")
)
