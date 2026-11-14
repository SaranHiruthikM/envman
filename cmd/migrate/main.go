// envman-migrate runs database migrations against DB_URL and exits. Deploy
// pipelines call this once per release rather than letting the server
// auto-migrate on boot.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	migrations "github.com/SaranHiruthikM/envman/migrations"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cmd := "up"
	flag.Parse()
	if flag.NArg() >= 1 {
		cmd = flag.Arg(0)
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		fmt.Fprintln(os.Stderr, "DB_URL not set")
		os.Exit(1)
	}

	driver, err := iofs.New(migrations.FS, ".")
	if err != nil {
		die(err)
	}
	pgxURL := strings.Replace(dbURL, "postgres://", "pgx5://", 1)
	m, err := migrate.NewWithSourceInstance("iofs", driver, pgxURL)
	if err != nil {
		die(err)
	}

	switch cmd {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "drop":
		err = m.Drop()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %q (use up|down|drop)\n", cmd)
		os.Exit(2)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		die(err)
	}
	slog.Info("migrate", "command", cmd, "ok", true)
}

func die(err error) {
	fmt.Fprintln(os.Stderr, "migrate:", err)
	os.Exit(1)
}
