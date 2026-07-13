package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 1 || (args[0] != "up" && args[0] != "status") {
		return fmt.Errorf("usage: migrate <up|status>")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "db/migrations"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	db, err := database.Open(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	migrator, err := database.NewMigrator(db, os.DirFS(migrationsDir), database.MigratorOptions{})
	if err != nil {
		return err
	}

	if args[0] == "up" {
		if err := migrator.Up(ctx); err != nil {
			return err
		}
	}

	statuses, err := migrator.Status(ctx)
	if err != nil {
		return err
	}
	for _, status := range statuses {
		fmt.Printf("%06d  %-8s  %s\n", status.Version, status.State, status.Name)
	}
	return nil
}
