package cmd

import (
	"errors"
	"net/http"

	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/mickaelfree/google-play-connect/internal/playapi"
)

// newStatusCmd reports active releases per track using the read-only
// release-summaries endpoint — no edit transaction involved.
func newStatusCmd(deps Deps, flags *RootFlags) *cobra.Command {
	var app, track string

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Show active releases per track (read-only, no edit)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			tracks := []string{track}
			if track == "" {
				tracks = []string{playapi.TrackInternal, playapi.TrackAlpha, playapi.TrackBeta, playapi.TrackProduction}
			}
			all := make([]*androidpublisher.ReleaseSummary, 0)
			for _, tr := range tracks {
				summaries, listErr := client.ListReleaseSummaries(cmd.Context(), app, tr)
				if listErr != nil {
					// When scanning all tracks, a 404 just means the track has
					// never had a release — skip it. Every other error (401,
					// 403, 500...) must surface, and an explicitly requested
					// track always surfaces its error.
					var apiErr *googleapi.Error
					if track == "" && errors.As(listErr, &apiErr) && apiErr.Code == http.StatusNotFound {
						continue
					}
					return listErr
				}
				all = append(all, summaries...)
			}
			return renderResult(deps, flags, all,
				[]string{"TRACK", "RELEASE", "STATE"},
				func() [][]string {
					rows := make([][]string, 0, len(all))
					for _, s := range all {
						rows = append(rows, []string{s.Track, s.ReleaseName, s.ReleaseLifecycleState})
					}
					return rows
				})
		},
	}
	statusCmd.Flags().StringVar(&app, "app", "", "package name of the app (required)")
	_ = statusCmd.MarkFlagRequired("app")
	statusCmd.Flags().StringVar(&track, "track", "", "restrict to one track (default: scan internal/alpha/beta/production)")

	return statusCmd
}
