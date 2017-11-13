package model

// Metadata is used by clients and server
// Before each event they send this to inform that the next event will be of type `EventType`
type Metadata struct {
	EventType string
}
