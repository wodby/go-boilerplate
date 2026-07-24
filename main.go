package main

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// webContent keeps the starter self-contained as a single deployable binary.
//
//go:embed web
var webContent embed.FS

var indexTemplate = template.Must(template.ParseFS(webContent, "web/index.html"))

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	address := ":" + port
	server := newServer(address, newHandler())
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	slog.Info("server starting", "address", address)
	if err := run(ctx, server); err != nil {
		log.Fatal(err)
	}
	slog.Info("server stopped")
}

func newServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func run(ctx context.Context, server *http.Server) error {
	errs := make(chan error, 1)
	go func() {
		errs <- server.ListenAndServe()
	}()

	select {
	case err := <-errs:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("shut down HTTP server: %w", err)
	}

	if err := <-errs; !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("serve HTTP during shutdown: %w", err)
	}
	return nil
}

func newHandler() http.Handler {
	mux := http.NewServeMux()
	assets, err := fs.Sub(webContent, "web/assets")
	if err != nil {
		panic(fmt.Sprintf("load embedded assets: %v", err))
	}

	mux.Handle(
		"GET /assets/",
		http.StripPrefix("/assets/", http.FileServer(http.FS(assets))),
	)
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		var body bytes.Buffer
		err := indexTemplate.ExecuteTemplate(
			&body,
			"index.html",
			map[string]string{"GoVersion": runtime.Version()},
		)
		if err != nil {
			http.Error(w, "render page", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(body.Bytes())
	})
	mux.HandleFunc("GET /api/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"runtime": runtime.Version(),
			"server":  "net/http",
		})
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = fmt.Fprintln(w, "ok")
	})
	return requestLogger(mux)
}

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, request)
		slog.Info(
			"request",
			"method", request.Method,
			"path", request.URL.Path,
			"duration", time.Since(started),
		)
	})
}
