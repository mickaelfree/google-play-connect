package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/mickaelfree/google-play-connect/internal/playapi"
)

func trackRows(tracks []*androidpublisher.Track) [][]string {
	rows := make([][]string, 0, len(tracks))
	for _, tr := range tracks {
		release := ""
		if len(tr.Releases) > 0 {
			r := tr.Releases[0]
			release = fmt.Sprintf("%s (%s)", r.Name, r.Status)
		}
		rows = append(rows, []string{tr.Track, release})
	}
	return rows
}

func newTracksCmd(deps Deps, flags *RootFlags) *cobra.Command {
	tracksCmd := &cobra.Command{
		Use:   "tracks",
		Short: "Inspect release tracks (internal, alpha, beta, production, custom)",
	}

	var app, track string

	list := &cobra.Command{
		Use:   "list",
		Short: "List all tracks and their current release",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var tracks []*androidpublisher.Track
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(id string) error {
				var inner error
				tracks, inner = client.ListTracks(cmd.Context(), app, id)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, tracks,
				[]string{"TRACK", "CURRENT_RELEASE"},
				func() [][]string { return trackRows(tracks) })
		},
	}

	get := &cobra.Command{
		Use:   "get",
		Short: "Show one track with all its releases",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var tr *androidpublisher.Track
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(id string) error {
				var inner error
				tr, inner = client.GetTrack(cmd.Context(), app, id, track)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, tr,
				[]string{"TRACK", "CURRENT_RELEASE"},
				func() [][]string { return trackRows([]*androidpublisher.Track{tr}) })
		},
	}
	get.Flags().StringVar(&track, "track", "", "track name (required)")
	_ = get.MarkFlagRequired("track")

	var editID, status string
	var rollout float64
	var versionCodes []int64
	var confirm bool
	update := &cobra.Command{
		Use:   "update",
		Short: "Patch a track's current release (rollout fraction, status, version codes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireConfirm(confirm, deps.Getenv, "update track"); err != nil {
				return err
			}
			if cmd.Flags().Changed("rollout") && (rollout <= 0 || rollout >= 1) {
				return fmt.Errorf("--rollout must be strictly between 0 and 1, got %v", rollout)
			}
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var updated *androidpublisher.Track
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				current, getErr := client.GetTrack(cmd.Context(), app, id, track)
				if getErr != nil {
					return getErr
				}
				if len(current.Releases) == 0 {
					return fmt.Errorf("track %s has no release to update; use gpc releases publish instead", track)
				}
				release := current.Releases[0]
				if cmd.Flags().Changed("rollout") {
					release.UserFraction = rollout
					release.Status = playapi.ReleaseStatusInProgress
				}
				if cmd.Flags().Changed("status") {
					release.Status = status
					if status == playapi.ReleaseStatusCompleted {
						release.UserFraction = 0
					}
				}
				if cmd.Flags().Changed("version-codes") {
					release.VersionCodes = googleapi.Int64s(versionCodes)
				}
				var inner error
				updated, inner = client.UpdateTrack(cmd.Context(), app, id, track, current)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, updated,
				[]string{"TRACK", "STATUS", "ROLLOUT"},
				func() [][]string {
					if len(updated.Releases) == 0 {
						return [][]string{{updated.Track, "-", "-"}}
					}
					r := updated.Releases[0]
					return [][]string{{updated.Track, r.Status, fmt.Sprint(r.UserFraction)}}
				})
		},
	}
	update.Flags().StringVar(&track, "track", "", "track name (required)")
	_ = update.MarkFlagRequired("track")
	update.Flags().Float64Var(&rollout, "rollout", 0, "new staged rollout fraction (0 < f < 1); sets status inProgress")
	update.Flags().StringVar(&status, "status", "", "new release status: inProgress, halted, or completed")
	update.Flags().Int64SliceVar(&versionCodes, "version-codes", nil, "replace the release's version codes")
	update.Flags().StringVar(&editID, "edit-id", "", "reuse an existing edit transaction (no auto-commit)")
	update.Flags().BoolVar(&confirm, "confirm", false, "confirm the track mutation")

	for _, c := range []*cobra.Command{list, get, update} {
		c.Flags().StringVar(&app, "app", "", "package name of the app (required)")
		_ = c.MarkFlagRequired("app")
	}

	tracksCmd.AddCommand(list, get, update)
	return tracksCmd
}
