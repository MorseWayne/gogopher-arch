package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
	"github.com/MorseWayne/gogopher-arch/internal/learning/projection"
	"github.com/MorseWayne/gogopher-arch/internal/platform/config"
	"github.com/MorseWayne/gogopher-arch/internal/platform/database"
)

type commandOptions struct {
	LearnerID    string
	CapabilityID string
	Apply        bool
	AsOf         time.Time
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:], os.Stdout, os.Stderr); err != nil {
		slog.Error("learning projection rebuild failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string, stdout, stderr io.Writer) error {
	options, err := parseOptions(args, stderr)
	if err != nil {
		return err
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	db, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	history, err := definition.NewReleaseStore(db, definition.ReleaseStoreOptions{})
	if err != nil {
		return err
	}
	registry, err := definition.BootstrapRegistry(ctx, cfg.LearningContentDir, history)
	if err != nil {
		return err
	}
	projector, err := projection.NewPostgresProjector(db, registry, projection.RepositoryOptions{})
	if err != nil {
		return err
	}
	targets, err := projector.RebuildTargets(ctx, options.LearnerID, options.CapabilityID, options.AsOf)
	if err != nil {
		return err
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	mode := "dry_run"
	if options.Apply {
		mode = "apply"
	}
	changed := 0
	for _, target := range targets {
		change, err := projector.Preview(ctx, target)
		if err != nil {
			return fmt.Errorf("preview %s %s@%d: %w", target.LearnerID, target.CapabilityID, target.CapabilityVersion, err)
		}
		if change.Change != "none" {
			changed++
			if options.Apply {
				applied, _, err := projector.Rebuild(ctx, target)
				if err != nil {
					return fmt.Errorf("apply %s %s@%d: %w", target.LearnerID, target.CapabilityID, target.CapabilityVersion, err)
				}
				change.After = applied
			}
		}
		if err := encoder.Encode(struct {
			Mode string `json:"mode"`
			projection.ProjectionChange
		}{Mode: mode, ProjectionChange: change}); err != nil {
			return err
		}
	}
	return encoder.Encode(struct {
		Mode    string `json:"mode"`
		Targets int    `json:"targets"`
		Changed int    `json:"changed"`
	}{Mode: mode, Targets: len(targets), Changed: changed})
}

func parseOptions(args []string, stderr io.Writer) (commandOptions, error) {
	flags := flag.NewFlagSet("learning-rebuild", flag.ContinueOnError)
	flags.SetOutput(stderr)
	learnerID := flags.String("learner-id", "", "rebuild one learner")
	capabilityID := flags.String("capability-id", "", "filter rebuild targets by Capability ID")
	dryRun := flags.Bool("dry-run", false, "preview changes without writing Snapshot or outbox")
	apply := flags.Bool("apply", false, "apply the displayed projection changes")
	if err := flags.Parse(args); err != nil {
		return commandOptions{}, err
	}
	if flags.NArg() != 0 {
		return commandOptions{}, fmt.Errorf("unexpected positional arguments")
	}
	if *dryRun && *apply {
		return commandOptions{}, errors.New("--dry-run and --apply are mutually exclusive")
	}
	return commandOptions{
		LearnerID: *learnerID, CapabilityID: *capabilityID,
		Apply: *apply, AsOf: time.Now().UTC(),
	}, nil
}
