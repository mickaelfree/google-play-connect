package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/mickaelfree/google-play-connect/internal/playapi"
)

// parseNotesFiles turns repeated "locale=path" values into LocalizedText by
// reading each file's contents.
func parseNotesFiles(specs []string) ([]*androidpublisher.LocalizedText, error) {
	notes := make([]*androidpublisher.LocalizedText, 0, len(specs))
	for _, spec := range specs {
		locale, path, ok := strings.Cut(spec, "=")
		if !ok {
			return nil, fmt.Errorf("invalid --notes-file %q: expected locale=path (e.g. en-US=notes/en.txt)", spec)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read notes file %s: %w", path, err)
		}
		notes = append(notes, &androidpublisher.LocalizedText{Language: locale, Text: strings.TrimSpace(string(data))})
	}
	return notes, nil
}

// notesFromDir reads every <locale>.txt file in dir as release notes.
func notesFromDir(dir string) ([]*androidpublisher.LocalizedText, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read notes dir %s: %w", dir, err)
	}
	var notes []*androidpublisher.LocalizedText
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		locale := strings.TrimSuffix(e.Name(), ".txt")
		data, readErr := os.ReadFile(filepath.Join(dir, e.Name()))
		if readErr != nil {
			return nil, fmt.Errorf("read notes file %s: %w", e.Name(), readErr)
		}
		notes = append(notes, &androidpublisher.LocalizedText{Language: locale, Text: strings.TrimSpace(string(data))})
	}
	return notes, nil
}

func newReleasesCmd(deps Deps, flags *RootFlags) *cobra.Command {
	releasesCmd := &cobra.Command{
		Use:   "releases",
		Short: "Publish releases to a track",
	}

	var app, editID, track, releaseName string
	var versionCodes []int64
	var rollout float64
	var notesDir string
	var notesFiles []string
	var confirm bool

	publish := &cobra.Command{
		Use:   "publish",
		Short: "Assign version codes to a track and roll out",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireConfirm(confirm, deps.Getenv, "publish release"); err != nil {
				return err
			}
			// rollout == 0 means "not set" (full rollout); anything else must
			// be a valid staged fraction.
			if rollout != 0 && (rollout <= 0 || rollout >= 1) {
				return fmt.Errorf("--rollout must be strictly between 0 and 1, got %v", rollout)
			}
			var notes []*androidpublisher.LocalizedText
			if notesDir != "" {
				dirNotes, dirErr := notesFromDir(notesDir)
				if dirErr != nil {
					return dirErr
				}
				notes = dirNotes
			}
			fileNotes, err := parseNotesFiles(notesFiles)
			if err != nil {
				return err
			}
			for _, n := range fileNotes {
				notes = playapi.UpsertReleaseNotes(notes, n.Language, n.Text)
			}

			status := playapi.ReleaseStatusCompleted
			var userFraction float64
			if rollout > 0 {
				status = playapi.ReleaseStatusInProgress
				userFraction = rollout
			}

			release := &androidpublisher.TrackRelease{
				Name:         releaseName,
				Status:       status,
				UserFraction: userFraction,
				VersionCodes: googleapi.Int64s(versionCodes),
				ReleaseNotes: notes,
			}
			trackBody := &androidpublisher.Track{
				Track:    track,
				Releases: []*androidpublisher.TrackRelease{release},
			}

			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var updated *androidpublisher.Track
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				var inner error
				updated, inner = client.UpdateTrack(cmd.Context(), app, id, track, trackBody)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, updated,
				[]string{"TRACK", "STATUS", "VERSION_CODES"},
				func() [][]string {
					r := updated.Releases[0]
					return [][]string{{updated.Track, r.Status, fmt.Sprint([]int64(r.VersionCodes))}}
				})
		},
	}
	publish.Flags().StringVar(&app, "app", "", "package name of the app (required)")
	_ = publish.MarkFlagRequired("app")
	publish.Flags().StringVar(&track, "track", "", "target track: internal, alpha, beta, production, or custom (required)")
	_ = publish.MarkFlagRequired("track")
	publish.Flags().Int64SliceVar(&versionCodes, "version-codes", nil, "version codes to release, e.g. --version-codes 42,43 (required)")
	_ = publish.MarkFlagRequired("version-codes")
	publish.Flags().StringVar(&releaseName, "name", "", "release name (defaults to version name server-side)")
	publish.Flags().Float64Var(&rollout, "rollout", 0, "staged rollout fraction (0 < f < 1); omit for full rollout")
	publish.Flags().StringVar(&notesDir, "notes-dir", "", "directory of <locale>.txt release notes (e.g. metadata/com.app/release_notes)")
	publish.Flags().StringArrayVar(&notesFiles, "notes-file", nil, "release notes as locale=path, repeatable")
	publish.Flags().StringVar(&editID, "edit-id", "", "reuse an existing edit transaction (no auto-commit)")
	publish.Flags().BoolVar(&confirm, "confirm", false, "confirm publishing")

	releasesCmd.AddCommand(publish)
	return releasesCmd
}
