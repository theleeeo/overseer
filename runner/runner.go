package runner

import (
	"context"
	"fmt"
	"log"
	"maps"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"overseer/app"
	"overseer/entrypoints"
	"overseer/repo"

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
	dbpool, err := pgxpool.New(ctx, r.config.DbConnString)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer dbpool.Close()

	queries := repo.New(dbpool)

	app := app.NewApp(queries)

	mux := http.NewServeMux()
	entrypoints.RegisterRestHandlers(mux, app)

	return http.ListenAndServe(":8080", LoggerMiddleware(mux))
}
