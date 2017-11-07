package event

// Metadata is used by clients and server
// Before each event they send this to inform that the next event will be of `EventType` type
type Metadata struct {
	EventType string
}
