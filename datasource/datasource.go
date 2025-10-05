package datasource

import "time"

type Event struct {
	Id             string // Unique identifier for the event
	DeploymentName string
	DeployedAt     time.Time
	Version        string
}

type EventStream <-chan Event
