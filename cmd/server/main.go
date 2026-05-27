package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/auth"
	"github.com/SaranHiruthikM/envman/internal/config"
	"github.com/SaranHiruthikM/envman/internal/db"
	"github.com/SaranHiruthikM/envman/internal/middleware"
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
	log.Println("Connected to database")

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

	log.Println("Migrations applied")

	mux := chi.NewMux()

	// Managers and Handlers
	tokenManager := token.NewManager(cfg.App.JWTSecret)
	authHandler := auth.NewHandler(pool, tokenManager)

	// Registering auth routes
	mux.Post("/auth/login", authHandler.Login)
	mux.Post("/auth/register", authHandler.Register)

	mux.Group(func(r chi.Router) {
		r.Use(middleware.RequireAuth(tokenManager))
		// protected routes go here
	})

	log.Printf("Server starting on port %s", cfg.App.Port)

	if err := http.ListenAndServe(fmt.Sprintf(":%s", cfg.App.Port), mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}

}
