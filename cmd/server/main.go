// Command server is the entrypoint for the Idea Maturity Platform API.
//
// With no DATABASE_URL set it runs against an in-memory store, which makes it
// trivial to start locally: `go run ./cmd/server`. When DATABASE_URL is set,
// wire the Postgres store here once a driver is added (see internal/store).
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/kezlya/nobel/internal/config"
	"github.com/kezlya/nobel/internal/httpapi"
	"github.com/kezlya/nobel/internal/store"
)

func main() {
	cfg := config.Load()

	var st store.Store
	if cfg.DatabaseURL == "" {
		log.Println("no DATABASE_URL set — using in-memory store (non-durable)")
		st = store.NewMemory()
	} else {
		// TODO: open *sql.DB with a registered Postgres driver and use
		// store.NewPostgres(db). Until then, fail loudly rather than silently
		// falling back to memory.
		log.Fatal("DATABASE_URL is set but the Postgres store is not wired yet; see internal/store/postgres.go")
	}

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           httpapi.New(st),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", cfg.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
