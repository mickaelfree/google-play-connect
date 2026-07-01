// Package cmd wires the gpc command tree. Commands never talk to the Google
// SDK directly: they go through internal/playapi via Deps.NewClient so tests
// can substitute a fake server.
package cmd

import (
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"
	androidpublisher "google.golang.org/api/androidpublisher/v3"

	"github.com/mickaelfree/google-play-connect/internal/auth"
	"github.com/mickaelfree/google-play-connect/internal/playapi"
)

// Deps carries the injectable seams for the command tree.
type Deps struct {
	// NewClient builds an authenticated Play API client. Tests replace it
	// with a stub pointed at a playapitest server.
	NewClient func(ctx context.Context, serviceAccountPath string) (*playapi.Client, error)
	// Stdout receives all rendered command output.
	Stdout io.Writer
	// Getenv reads environment variables (credential fallbacks, CI detection).
	Getenv func(string) string
}

// RootFlags holds the persistent flags shared by every subcommand.
type RootFlags struct {
	ServiceAccount string
	Output         string
}

// DefaultDeps returns the production wiring used by main().
func DefaultDeps() Deps {
	return Deps{
		Stdout: os.Stdout,
		Getenv: os.Getenv,
		NewClient: func(ctx context.Context, serviceAccountPath string) (*playapi.Client, error) {
			creds, err := auth.ResolveCredentials(auth.Config{ServiceAccountPath: serviceAccountPath}, os.Getenv)
			if err != nil {
				return nil, err
			}
			opts, err := creds.ClientOptions()
			if err != nil {
				return nil, err
			}
			svc, err := androidpublisher.NewService(ctx, opts...)
			if err != nil {
				return nil, err
			}
			return playapi.NewClient(svc), nil
		},
	}
}

// NewRootCmd builds the gpc command tree.
func NewRootCmd(deps Deps) *cobra.Command {
	flags := &RootFlags{}
	root := &cobra.Command{
		Use:           "gpc",
		Short:         "Google Play Connect — manage Google Play Console from the terminal",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.PersistentFlags().StringVar(&flags.ServiceAccount, "service-account", "", "path to the service account JSON key")
	root.PersistentFlags().StringVar(&flags.Output, "output", "", "output format: json or table (default: json, table on a TTY)")
	root.AddCommand(
		newAppsCmd(deps, flags),
		newEditsCmd(deps, flags),
		newListingsCmd(deps, flags),
		newImagesCmd(deps, flags),
		newTracksCmd(deps, flags),
		newReleasesCmd(deps, flags),
	)
	return root
}

// Execute runs the CLI with production dependencies.
func Execute() {
	if err := NewRootCmd(DefaultDeps()).Execute(); err != nil {
		os.Exit(1)
	}
}
