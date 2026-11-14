package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/SaranHiruthikM/envman/internal/audit"
	"github.com/SaranHiruthikM/envman/internal/auth"
	"github.com/SaranHiruthikM/envman/internal/config"
	"github.com/SaranHiruthikM/envman/internal/db"
	"github.com/SaranHiruthikM/envman/internal/members"
	appmw "github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/projects"
	"github.com/SaranHiruthikM/envman/internal/secrets"
	"github.com/SaranHiruthikM/envman/internal/token"
	"github.com/SaranHiruthikM/envman/internal/users"
	migrations "github.com/SaranHiruthikM/envman/migrations"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()
	slog.Info("database connected")

	if os.Getenv("RUN_MIGRATIONS") == "true" {
		if err := runMigrations(cfg.DB.DBUrl); err != nil {
			slog.Error("migrations failed", "err", err)
			os.Exit(1)
		}
		slog.Info("migrations applied")
	}

	tokenMgr := token.NewManager(cfg.App.JWTSecret)
	authHandler := auth.NewHandler(pool, tokenMgr)
	memberHandler := members.NewHandler(pool)
	secretHandler := secrets.NewHandler(pool)
	auditHandler := audit.NewHandler(pool)
	projectHandler := projects.NewHandler(pool, memberHandler, secretHandler, auditHandler)
	usersHandler := users.NewHandler(pool)

	mux := chi.NewMux()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
	mux.Use(appmw.LimitBody)
	mux.Use(requestLogger)

	mux.Get("/healthz", healthz)
	mux.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := pool.Ping(r.Context()); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Public
	mux.Post("/auth/login", authHandler.Login)
	mux.Post("/auth/register", authHandler.Register)

	// Authenticated
	mux.Group(func(r chi.Router) {
		r.Use(appmw.RequireAuth(tokenMgr))
		r.Mount("/projects", projectHandler.Routes())
		r.Mount("/users", usersHandler.Routes())
	})

	srv := &http.Server{
		Addr:              ":" + cfg.App.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received")
	case err := <-serverErr:
		slog.Error("server error", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown failed", "err", err)
	}
	slog.Info("stopped")
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// requestLogger emits one structured line per request. Status codes from
// handlers are captured via a wrapped ResponseWriter.
func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		// Skip noisy health checks
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			return
		}
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"bytes", ww.BytesWritten(),
			"duration_ms", time.Since(start).Milliseconds(),
			"request_id", middleware.GetReqID(r.Context()),
		)
	})
}

// runMigrations applies all pending migrations from the embedded SQL files.
func runMigrations(dbURL string) error {
	driver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}
	pgxURL := strings.Replace(dbURL, "postgres://", "pgx5://", 1)
	m, err := migrate.NewWithSourceInstance("iofs", driver, pgxURL)
	if err != nil {
		return err
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
