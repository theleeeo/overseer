package datasource

type Event struct {
	Id string // Unique identifier for the event
}

type EventStream <-chan Event
