package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"

	"github.com/mickaelfree/google-play-connect/internal/metadata"
	"github.com/mickaelfree/google-play-connect/internal/playapi"
)

func newMetadataCmd(deps Deps, flags *RootFlags) *cobra.Command {
	metadataCmd := &cobra.Command{
		Use:   "metadata",
		Short: "Pull, validate and push the store listing tree (offline editing)",
	}

	var app, dir, editID string
	var confirm, prune bool

	pull := &cobra.Command{
		Use:   "pull",
		Short: "Download listings and image manifests into --dir",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var pulledLocales []string
			// writtenManifests tracks which locale/imageType manifests this
			// pull (re)wrote, so --prune only removes manifests it did not
			// touch (i.e. ones that no longer exist remotely).
			writtenManifests := map[string]map[string]bool{}
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(id string) error {
				listings, inner := client.ListListings(cmd.Context(), app, id)
				if inner != nil {
					return inner
				}
				for _, l := range listings {
					if writeErr := metadata.WriteListing(dir, app, l.Language, metadata.Listing{
						Title:            l.Title,
						ShortDescription: l.ShortDescription,
						FullDescription:  l.FullDescription,
						Video:            l.Video,
					}); writeErr != nil {
						return writeErr
					}
					pulledLocales = append(pulledLocales, l.Language)
					for _, imageType := range playapi.AllImageTypes {
						images, listErr := client.ListImages(cmd.Context(), app, id, l.Language, imageType)
						if listErr != nil {
							return listErr
						}
						if len(images) == 0 {
							continue
						}
						entries := make([]metadata.ImageManifestEntry, 0, len(images))
						for _, img := range images {
							entries = append(entries, metadata.ImageManifestEntry{ID: img.Id, Sha1: img.Sha1, URL: img.Url})
						}
						if mErr := metadata.WriteImageManifest(dir, app, l.Language, imageType, entries); mErr != nil {
							return mErr
						}
						if writtenManifests[l.Language] == nil {
							writtenManifests[l.Language] = map[string]bool{}
						}
						writtenManifests[l.Language][imageType] = true
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			prunedPaths := []string{}
			if prune {
				prunedPaths, err = pruneStaleMetadata(dir, app, pulledLocales, writtenManifests)
				if err != nil {
					return err
				}
			}
			return renderResult(deps, flags, map[string]any{"pulled": pulledLocales, "dir": dir, "pruned": prunedPaths},
				[]string{"LOCALE"},
				func() [][]string {
					rows := make([][]string, 0, len(pulledLocales))
					for _, l := range pulledLocales {
						rows = append(rows, []string{l})
					}
					return rows
				})
		},
	}
	pull.Flags().BoolVar(&prune, "prune", false, "delete local listing locales and image manifests that no longer exist remotely (local image files are never touched)")

	push := &cobra.Command{
		Use:   "push",
		Short: "Apply the local tree (listings + images) in one transaction",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireConfirm(confirm, deps.Getenv, "metadata push"); err != nil {
				return err
			}
			issues, err := metadata.ValidateTree(dir, app)
			if err != nil {
				return err
			}
			if len(issues) > 0 {
				_ = renderResult(deps, flags, issues, nil, nil)
				return fmt.Errorf("metadata validation failed with %d issue(s); fix them or run gpc metadata validate", len(issues))
			}

			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var pushedLocales []string
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				locales, inner := metadata.ListingLocales(dir, app)
				if inner != nil {
					return inner
				}
				for _, locale := range locales {
					l, readErr := metadata.ReadListing(dir, app, locale)
					if readErr != nil {
						return readErr
					}
					if _, upErr := client.UpdateListing(cmd.Context(), app, id, locale, &androidpublisher.Listing{
						Language:         locale,
						Title:            l.Title,
						ShortDescription: l.ShortDescription,
						FullDescription:  l.FullDescription,
						Video:            l.Video,
					}); upErr != nil {
						return upErr
					}
					pushedLocales = append(pushedLocales, locale)
				}
				imageLocales, inner := metadata.ImageLocales(dir, app)
				if inner != nil {
					return inner
				}
				for _, locale := range imageLocales {
					for _, imageType := range playapi.AllImageTypes {
						paths, listErr := metadata.LocalImages(dir, app, locale, imageType)
						if listErr != nil {
							return listErr
						}
						if len(paths) == 0 {
							continue
						}
						if _, delErr := client.DeleteAllImages(cmd.Context(), app, id, locale, imageType); delErr != nil {
							return delErr
						}
						for _, path := range paths {
							f, openErr := os.Open(path)
							if openErr != nil {
								return fmt.Errorf("open %s: %w", path, openErr)
							}
							_, upErr := client.UploadImage(cmd.Context(), app, id, locale, imageType, f)
							f.Close()
							if upErr != nil {
								return upErr
							}
						}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, map[string]any{"pushed": pushedLocales},
				[]string{"LOCALE"},
				func() [][]string {
					rows := make([][]string, 0, len(pushedLocales))
					for _, l := range pushedLocales {
						rows = append(rows, []string{l})
					}
					return rows
				})
		},
	}
	push.Flags().BoolVar(&confirm, "confirm", false, "confirm applying local metadata to Play Console")
	push.Flags().StringVar(&editID, "edit-id", "", "reuse an existing edit transaction (no auto-commit)")

	validate := &cobra.Command{
		Use:   "validate",
		Short: "Check the local tree against Play limits (offline, no API calls)",
		RunE: func(cmd *cobra.Command, args []string) error {
			issues, err := metadata.ValidateTree(dir, app)
			if err != nil {
				return err
			}
			renderErr := renderResult(deps, flags, issues,
				[]string{"LOCALE", "FIELD", "MESSAGE"},
				func() [][]string {
					rows := make([][]string, 0, len(issues))
					for _, i := range issues {
						rows = append(rows, []string{i.Locale, i.Field, i.Message})
					}
					return rows
				})
			if renderErr != nil {
				return renderErr
			}
			if len(issues) > 0 {
				return fmt.Errorf("%d validation issue(s)", len(issues))
			}
			return nil
		},
	}

	for _, c := range []*cobra.Command{pull, push, validate} {
		c.Flags().StringVar(&app, "app", "", "package name of the app (required)")
		_ = c.MarkFlagRequired("app")
		c.Flags().StringVar(&dir, "dir", "./metadata", "root of the metadata tree")
	}

	metadataCmd.AddCommand(pull, push, validate)
	return metadataCmd
}

// pruneStaleMetadata removes local listing locale directories and image
// manifests that this pull did not (re)write, i.e. no longer exist remotely.
// It never touches image files or image directories. Returned paths are
// relative to dir.
func pruneStaleMetadata(dir, app string, pulledLocales []string, writtenManifests map[string]map[string]bool) ([]string, error) {
	pruned := []string{}

	localListingLocales, err := metadata.ListingLocales(dir, app)
	if err != nil {
		return nil, err
	}
	for _, locale := range localListingLocales {
		if slices.Contains(pulledLocales, locale) {
			continue
		}
		listingDir := filepath.Join(metadata.AppDir(dir, app), "listings", locale)
		if rmErr := os.RemoveAll(listingDir); rmErr != nil {
			return nil, fmt.Errorf("prune %s: %w", listingDir, rmErr)
		}
		pruned = append(pruned, filepath.Join(app, "listings", locale))
	}

	localImageLocales, err := metadata.ImageLocales(dir, app)
	if err != nil {
		return nil, err
	}
	for _, locale := range localImageLocales {
		for _, imageType := range playapi.AllImageTypes {
			manifestPath := filepath.Join(metadata.AppDir(dir, app), "images", locale, imageType+".manifest.json")
			if _, statErr := os.Stat(manifestPath); statErr != nil {
				continue
			}
			if writtenManifests[locale][imageType] {
				continue
			}
			if rmErr := os.Remove(manifestPath); rmErr != nil {
				return nil, fmt.Errorf("prune %s: %w", manifestPath, rmErr)
			}
			pruned = append(pruned, filepath.Join(app, "images", locale, imageType+".manifest.json"))
		}
	}

	return pruned, nil
}
