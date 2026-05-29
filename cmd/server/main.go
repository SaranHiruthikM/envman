package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/audit"
	"github.com/SaranHiruthikM/envman/internal/auth"
	"github.com/SaranHiruthikM/envman/internal/config"
	"github.com/SaranHiruthikM/envman/internal/db"
	"github.com/SaranHiruthikM/envman/internal/members"
	"github.com/SaranHiruthikM/envman/internal/middleware"
	"github.com/SaranHiruthikM/envman/internal/projects"
	"github.com/SaranHiruthikM/envman/internal/secrets"
	"github.com/SaranHiruthikM/envman/internal/token"
	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	// New DB Pool Creation
	pool, err := db.NewPool(ctx, cfg.DB)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// Migration Connection Code
	migrateUrl := strings.Replace(cfg.DB.DBUrl, "postgres://", "pgx5://", 1)
	m, err := migrate.New("file://migrations", migrateUrl)
	if err != nil {
		log.Fatalf("Error running migrations: %v", err)
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Error running migrations: %v", err)
	}

	mux := chi.NewMux()

	// Managers and Handlers
	tokenManager := token.NewManager(cfg.App.JWTSecret)
	authHandler := auth.NewHandler(pool, tokenManager)
	memberHandler := members.NewHandler(pool)
	secretHandler := secrets.NewHandler(pool)
	logsHandler := audit.NewHandler(pool)
	projectHandler := projects.NewHandler(pool, memberHandler, secretHandler, logsHandler)

	// Registering auth routes (Public routes)
	mux.Post("/auth/login", authHandler.Login)
	mux.Post("/auth/register", authHandler.Register)

	// Projects Routes
	mux.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(tokenManager))
		r.Mount("/projects", projectHandler.Routes())
	})

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.App.Port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
