package runner

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"overseer/app"
	"overseer/datasource"
	"overseer/entrypoints"
	"overseer/repo"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DbConnString string
}

type Runner struct {
	config *Config
}

func New(config *Config) *Runner {
	return &Runner{config: config}
}

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		respCatcher := httptest.NewRecorder()
		next.ServeHTTP(respCatcher, r)

		if respCatcher.Code == http.StatusInternalServerError {
			responseId := rand.Intn(1000000) //nolint:gosec // This is not for security purposes, it does not have to be cryptographically secure

			log.Printf("internal error: %s, id: %d, path: %s", removeTrailingNewline(respCatcher.Body.String()), responseId, r.URL.Path)

			respCatcher.Body.Reset()
			_, err := respCatcher.Body.WriteString(fmt.Sprintf("internal error, id: %d", responseId))
			if err != nil {
				log.Printf("failed to write to response body: %v", err)
			}
		}

		log.Printf("%s %s %d", r.Method, r.URL.Path, respCatcher.Code)

		maps.Copy(w.Header(), respCatcher.Header())

		w.WriteHeader(respCatcher.Code)
		_, _ = w.Write(respCatcher.Body.Bytes())
	})
}

func removeTrailingNewline(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}

func (r *Runner) Run(ctx context.Context) error {
	termChan := make(chan os.Signal, 1)
	errChan := make(chan error, 1)
	signal.Notify(termChan, os.Interrupt)

	dbpool, err := pgxpool.New(ctx, r.config.DbConnString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbpool.Close()

	queries := repo.New(dbpool)

	app := app.NewApp(queries)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	nomadSource := datasource.NewNomadSource("http://localhost:4646", "your-nomad-token")
	eventStream, err := nomadSource.StreamEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to start event stream: %w", err)
	}

	mux := http.NewServeMux()
	entrypoints.RegisterRestHandlers(mux, app)

	server := &http.Server{Addr: ":8080", Handler: LoggerMiddleware(mux)}

	wg := sync.WaitGroup{}
	wg.Go(func() {
		if err := app.RunVersionStream(eventStream); err != nil {
			errChan <- fmt.Errorf("version stream error: %w", err)
		}

		slog.Info("Event stream processing stopped.")
	})

	wg.Go(func() {
		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				errChan <- fmt.Errorf("server shutdown error: %w", err)
			}
		}()

		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("http server error: %w", err)
		}

		slog.Info("HTTP server stopped.")
	})

	select {
	case <-termChan:
		log.Println("Received termination signal, shutting down...")
	case err := <-errChan:
		log.Printf("Received error: %v, shutting down...", err)
	}

	go func() {
		for err := range errChan {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	go func() {
		<-termChan
		log.Println("Received second termination signal, forcing shutdown...")
		os.Exit(1)
	}()

	cancel()
	wg.Wait()
	log.Println("Shutdown complete.")

	return nil
}
