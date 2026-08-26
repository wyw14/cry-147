package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/wyw14/cry-147/internal/api"
	cellruntime "github.com/wyw14/cry-147/internal/runtime"
	"github.com/wyw14/cry-147/internal/service"
)

func env(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func main() {
	address := env("CELLFORGE_ADDR", "127.0.0.1:21247")
	dataDir := env("CELLFORGE_DATA", filepath.Join(os.TempDir(), "cellforge"))
	webRoot := env("CELLFORGE_WEB", "web")
	thermalURL := env("CELLFORGE_THERMAL_URL", "http://127.0.0.1:21248")
	system, err := cellruntime.New(dataDir, thermalURL)
	if err != nil {
		log.Fatal(err)
	}
	runtimeService := service.NewRuntime(system)
	server := &http.Server{Addr: address, Handler: api.NewServer(runtimeService, webRoot).Handler(), ReadHeaderTimeout: 5 * time.Second}
	errorsChannel := make(chan error, 1)
	go func() {
		log.Printf("CellForge listening on http://%s", address)
		errorsChannel <- server.ListenAndServe()
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-signals:
	case err := <-errorsChannel:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown failed: %v", err)
	}
}
