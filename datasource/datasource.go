package datasource

type Event struct {
	Id string // Unique identifier for the event
	// ResumeToken string // Token to resume event streaming from this point
}

type EventStream <-chan Event
