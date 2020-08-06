package v1

type Status string

const (
	StatusSucceeded = Status("Succeeded")
	StatusReady     = Status("Ready")
	StatusFailed    = Status("Failed")
	StatusRunning   = Status("Running")
	StatusPending   = Status("Pending")
	StatusSkipped   = Status("Skipped")
)
