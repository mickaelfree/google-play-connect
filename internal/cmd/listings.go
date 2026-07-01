package cmd

import (
	"errors"
	"net/http"

	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func newListingsCmd(deps Deps, flags *RootFlags) *cobra.Command {
	listingsCmd := &cobra.Command{
		Use:   "listings",
		Short: "Manage per-locale store listings (title, descriptions, video)",
	}

	var app, editID, locale string

	list := &cobra.Command{
		Use:   "list",
		Short: "List every locale's store listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var listings []*androidpublisher.Listing
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(id string) error {
				var inner error
				listings, inner = client.ListListings(cmd.Context(), app, id)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, listings,
				[]string{"LOCALE", "TITLE", "SHORT_DESCRIPTION"},
				func() [][]string {
					rows := make([][]string, 0, len(listings))
					for _, l := range listings {
						rows = append(rows, []string{l.Language, l.Title, l.ShortDescription})
					}
					return rows
				})
		},
	}

	get := &cobra.Command{
		Use:   "get",
		Short: "Show one locale's store listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var listing *androidpublisher.Listing
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(id string) error {
				var inner error
				listing, inner = client.GetListing(cmd.Context(), app, id, locale)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, listing,
				[]string{"LOCALE", "TITLE", "SHORT_DESCRIPTION", "VIDEO"},
				func() [][]string {
					return [][]string{{listing.Language, listing.Title, listing.ShortDescription, listing.Video}}
				})
		},
	}

	var title, shortDesc, fullDesc, video string
	update := &cobra.Command{
		Use:   "update",
		Short: "Create or update one locale's store listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var updated *androidpublisher.Listing
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				existing, getErr := client.GetListing(cmd.Context(), app, id, locale)
				if getErr != nil {
					var apiErr *googleapi.Error
					if errors.As(getErr, &apiErr) && apiErr.Code == http.StatusNotFound {
						// New locale: start from an empty listing.
						existing = &androidpublisher.Listing{Language: locale}
					} else {
						return getErr
					}
				}
				next := &androidpublisher.Listing{
					Language:         locale,
					Title:            existing.Title,
					ShortDescription: existing.ShortDescription,
					FullDescription:  existing.FullDescription,
					Video:            existing.Video,
				}
				if cmd.Flags().Changed("title") {
					next.Title = title
				}
				if cmd.Flags().Changed("short-description") {
					next.ShortDescription = shortDesc
				}
				if cmd.Flags().Changed("full-description") {
					next.FullDescription = fullDesc
				}
				if cmd.Flags().Changed("video") {
					next.Video = video
				}
				var inner error
				updated, inner = client.UpdateListing(cmd.Context(), app, id, locale, next)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, updated,
				[]string{"LOCALE", "TITLE"},
				func() [][]string { return [][]string{{updated.Language, updated.Title}} })
		},
	}
	update.Flags().StringVar(&title, "title", "", "app title (max 30 chars)")
	update.Flags().StringVar(&shortDesc, "short-description", "", "short description (max 80 chars)")
	update.Flags().StringVar(&fullDesc, "full-description", "", "full description (max 4000 chars)")
	update.Flags().StringVar(&video, "video", "", "promo YouTube video URL")

	var confirm bool
	del := &cobra.Command{
		Use:   "delete",
		Short: "Delete one locale's store listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireConfirm(confirm, deps.Getenv, "delete listing"); err != nil {
				return err
			}
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				return client.DeleteListing(cmd.Context(), app, id, locale)
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, map[string]string{"deleted": locale},
				[]string{"DELETED"}, func() [][]string { return [][]string{{locale}} })
		},
	}
	del.Flags().BoolVar(&confirm, "confirm", false, "confirm the destructive action")

	for _, c := range []*cobra.Command{list, get, update, del} {
		c.Flags().StringVar(&app, "app", "", "package name of the app (required)")
		_ = c.MarkFlagRequired("app")
	}
	for _, c := range []*cobra.Command{get, update, del} {
		c.Flags().StringVar(&locale, "locale", "", "BCP-47 locale, e.g. fr-FR (required)")
		_ = c.MarkFlagRequired("locale")
	}
	for _, c := range []*cobra.Command{update, del} {
		c.Flags().StringVar(&editID, "edit-id", "", "reuse an existing edit transaction (no auto-commit)")
	}

	listingsCmd.AddCommand(list, get, update, del)
	return listingsCmd
}
