package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/MorseWayne/gogopher-arch/internal/learning/definition"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: learning-content <validate|release|verify> [flags]")
	}
	switch args[0] {
	case "validate":
		flags := flag.NewFlagSet("validate", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		contentDir := flags.String("content-dir", "content/learning", "draft content root")
		activitySet := flags.String("activity-set", "", "activity set directory name")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *activitySet == "" {
			return fmt.Errorf("validate requires --activity-set and accepts no positional arguments")
		}
		if err := definition.ValidateDrafts(*contentDir, *activitySet); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "validated activity set %s\n", *activitySet)
		return nil

	case "release":
		flags := flag.NewFlagSet("release", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		contentDir := flags.String("content-dir", "content/learning", "draft content root")
		activitySet := flags.String("activity-set", "", "activity set directory name")
		releaseID := flags.String("release-id", "", "immutable release identifier")
		createdAtValue := flags.String("created-at", "", "RFC3339 release creation time")
		outputDir := flags.String("output-dir", "", "release parent directory")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *activitySet == "" || *releaseID == "" || *createdAtValue == "" {
			return fmt.Errorf("release requires --activity-set, --release-id, and --created-at")
		}
		createdAt, err := time.Parse(time.RFC3339, *createdAtValue)
		if err != nil {
			return fmt.Errorf("parse --created-at: %w", err)
		}
		target, err := definition.BuildRelease(definition.ReleaseOptions{
			ContentDir: *contentDir, ActivitySet: *activitySet, ReleaseID: *releaseID,
			CreatedAt: createdAt, OutputDir: *outputDir,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "released %s\n", target)
		return nil

	case "verify":
		flags := flag.NewFlagSet("verify", flag.ContinueOnError)
		flags.SetOutput(io.Discard)
		releaseDir := flags.String("release-dir", "", "release directory containing manifest.json")
		schemasDir := flags.String("schemas-dir", "", "JSON schema directory")
		webDist := flags.String("web-dist", "", "optional frontend build directory to inspect for held-out assets")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		if flags.NArg() != 0 || *releaseDir == "" {
			return fmt.Errorf("verify requires --release-dir and accepts no positional arguments")
		}
		if *schemasDir == "" {
			*schemasDir = filepath.Join(filepath.Dir(filepath.Dir(filepath.Clean(*releaseDir))), "schemas")
		}
		if err := definition.VerifyRelease(*releaseDir, *schemasDir); err != nil {
			return err
		}
		if *webDist != "" {
			if err := definition.VerifyFrontendBundle(*releaseDir, *webDist); err != nil {
				return err
			}
		}
		fmt.Fprintf(stdout, "verified %s\n", *releaseDir)
		return nil

	default:
		return fmt.Errorf("unknown command %q; expected validate, release, or verify", args[0])
	}
}
