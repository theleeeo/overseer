package nomad

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"overseer/datasource"
	"strings"
	"time"
)

type Source struct {
	address string
	token   string
	logger  *slog.Logger
}

type streamItem struct {
	Index  int64
	Events []event
}

type event struct {
	Topic      string
	Type       string
	Key        string
	Namespace  string
	FilterKeys []string
	Index      uint64
	Payload    json.RawMessage
}

type job struct {
	Namespace  string
	ID         string
	Name       string // Whats the difference to ID?
	TaskGroups []taskGroup
	SubmitTime int64 // Unix timestamp in nanoseconds
	// Meta       map[string]any
}

type taskGroup struct {
	Name  string
	Tasks []task
}

type task struct {
	Name   string
	Driver string
	Config json.RawMessage
}

type dockerConfig struct {
	Image string // Might need a JSON tag for lowercase
}

// TODO: Should handle the initial sync of existing jobs.
func NewSource(address, token string, logger *slog.Logger) *Source {
	address = address + "/v1/event/stream?topic=Job"

	return &Source{
		address: address,
		token:   token,
		logger:  logger,
	}
}

func (n *Source) StreamEvents(ctx context.Context) (<-chan datasource.Event, error) {
	stream := make(chan datasource.Event, 10)

	go func() {
		defer close(stream)

		for {
			err := n.runStream(ctx, stream)
			if errors.Is(err, context.Canceled) {
				return
			}

			n.logger.Error("stream error", "error", err)

			// TODO: Sleep with backoff?
			time.Sleep(5 * time.Second)

			n.logger.Info("reconnecting to Nomad event stream...")
		}
	}()

	return stream, nil
}

func (n *Source) runStream(ctx context.Context, stream chan datasource.Event) error {
	b, err := n.getNomadConnection(ctx)
	if err != nil {
		return err
	}
	defer b.Close()

	decoder := json.NewDecoder(b)

	for {
		var streamItem streamItem
		if err := decoder.Decode(&streamItem); err != nil {
			return err
		}

		// n.logger.Info("received stream item", "index", streamItem.Index, "events", len(streamItem.Events))

		for _, event := range streamItem.Events {
			if event.Type != "JobRegistered" { // TODO: For now, only handle JobRegistered
				continue
			}

			var jp struct {
				Job job
			}
			if err := json.Unmarshal(event.Payload, &jp); err != nil {
				return fmt.Errorf("parsing job payload: %w", err)
			}
			j := jp.Job

			for _, tg := range j.TaskGroups {
				for _, task := range tg.Tasks {
					if task.Driver != "docker" {
						continue
					}

					var config dockerConfig
					if err := json.Unmarshal(task.Config, &config); err != nil {
						n.logger.Error("parsing docker config", "error", err)
						continue
					}

					var imageVersion string
					lastColon := strings.LastIndex(config.Image, ":")
					if lastColon != -1 && lastColon < len(config.Image)-1 {
						imageVersion = config.Image[lastColon+1:]
					} else {
						imageVersion = ""
					}

					if imageVersion == "" || imageVersion == "latest" {
						n.logger.Warn("could not determine image version", "image", config.Image)
						continue
					}

					stream <- datasource.Event{
						Id:             event.Key,
						DeploymentName: fmt.Sprintf("%s.%s.%s.%s", j.Namespace, j.Name, tg.Name, task.Name),
						Version:        imageVersion,
						DeployedAt:     time.Unix(0, j.SubmitTime),
					}
				}
			}
		}
	}
}

func (n *Source) getNomadConnection(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", n.address, nil)
	if err != nil {
		return nil, fmt.Errorf("creating the request: %w", err)
	}
	req.Header.Set("X-Nomad-Token", n.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connecting to Nomad: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status from Nomad: %s", resp.Status)
	}

	return resp.Body, nil
}
