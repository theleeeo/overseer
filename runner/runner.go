package runner

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	applicationpb "overseer/api-go/application/v1"
	deploymentpb "overseer/api-go/deployment/v1"
	environmentpb "overseer/api-go/environment/v1"
	instancepb "overseer/api-go/instance/v1"
	"overseer/app"
	"overseer/datasource"
	"overseer/entrypoints"
	"overseer/repo"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
			_, err := fmt.Fprintf(respCatcher.Body, "internal error, id: %d", responseId)
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

	app := app.New(queries)

	mux := http.NewServeMux()
	entrypoints.RegisterRestHandlers(mux, app)
	server := &http.Server{Addr: ":8080", Handler: LoggerMiddleware(mux)}

	applicationGrpc := entrypoints.NewApplicationServer(app)
	environmentGrpc := entrypoints.NewEnvironmentServer(app)
	deploymentGrpc := entrypoints.NewDeploymentServer(app)
	instanceGrpc := entrypoints.NewInstanceServer(app)

	grpcServer := grpc.NewServer(
	// grpc.MaxRecvMsgSize(mb256),
	// grpc.MaxSendMsgSize(mb256),
	// grpc.ChainUnaryInterceptor(unaryInterceptors...),
	)

	reflection.Register(grpcServer)

	applicationpb.RegisterApplicationServiceServer(grpcServer, applicationGrpc)
	environmentpb.RegisterEnvironmentServiceServer(grpcServer, environmentGrpc)
	deploymentpb.RegisterDeploymentServiceServer(grpcServer, deploymentGrpc)
	instancepb.RegisterInstanceServiceServer(grpcServer, instanceGrpc)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// dataSource := nomad.NewSource("http://rock-srv-1.local:4646", "your-nomad-token", slog.Default())
	dataSource := &datasource.MockSource{}

	wg := sync.WaitGroup{}

	wg.Go(func() {
		if err := app.RunVersionStream(ctx, dataSource); err != nil {
			errChan <- fmt.Errorf("version stream error: %w", err)
		}

		slog.Info("Event stream processing stopped.")
	})

	wg.Go(func() {
		defer slog.Info("gRPC server stopped")

		go func() {
			<-ctx.Done()
			grpcServer.GracefulStop()
		}()

		slog.Info("starting the grpc server", slog.String("address", "localhost:9090"))

		lis, err := net.Listen("tcp", "localhost:9090")
		if err != nil {
			errChan <- fmt.Errorf("failed to listen on port 9090: %w", err)
			return
		}

		if err := grpcServer.Serve(lis); err != nil {
			errChan <- fmt.Errorf("grpc server error: %w", err)
		}

	})

	wg.Go(func() {
		defer slog.Info("HTTP server stopped.")

		go func() {
			<-ctx.Done()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			if err := server.Shutdown(shutdownCtx); err != nil {
				errChan <- fmt.Errorf("server shutdown error: %w", err)
			}
		}()

		slog.Info("starting the http server on :8080")

		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			errChan <- fmt.Errorf("http server error: %w", err)
		}
	})

	select {
	case <-termChan:
		log.Println("Received termination signal")
	case err := <-errChan:
		log.Printf("Received error: %v", err)
	}

	log.Println("Shutting down...")

	cancel()

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

	wg.Wait()
	log.Println("Shutdown complete.")

	return nil
}
