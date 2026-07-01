package cmd

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"

	"github.com/mickaelfree/google-play-connect/internal/playapi"
)

func validateImageType(imageType string) error {
	if slices.Contains(playapi.AllImageTypes, imageType) {
		return nil
	}
	return fmt.Errorf("invalid --type %q: must be one of %v", imageType, playapi.AllImageTypes)
}

func newImagesCmd(deps Deps, flags *RootFlags) *cobra.Command {
	imagesCmd := &cobra.Command{
		Use:   "images",
		Short: "Manage store images (screenshots, icon, feature graphic)",
	}

	var app, editID, locale, imageType string
	var confirm bool

	list := &cobra.Command{
		Use:   "list",
		Short: "List images for a locale and image type",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateImageType(imageType); err != nil {
				return err
			}
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var images []*androidpublisher.Image
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(id string) error {
				var inner error
				images, inner = client.ListImages(cmd.Context(), app, id, locale, imageType)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, images,
				[]string{"ID", "URL"},
				func() [][]string {
					rows := make([][]string, 0, len(images))
					for _, img := range images {
						rows = append(rows, []string{img.Id, img.Url})
					}
					return rows
				})
		},
	}

	upload := &cobra.Command{
		Use:   "upload <file.png> [more files...]",
		Short: "Upload one or more images to a locale and image type",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateImageType(imageType); err != nil {
				return err
			}
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var uploaded []*androidpublisher.Image
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				for _, path := range args {
					f, openErr := os.Open(path)
					if openErr != nil {
						return fmt.Errorf("open %s: %w", path, openErr)
					}
					img, upErr := client.UploadImage(cmd.Context(), app, id, locale, imageType, f)
					f.Close()
					if upErr != nil {
						return upErr
					}
					uploaded = append(uploaded, img)
				}
				return nil
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, uploaded,
				[]string{"ID", "URL"},
				func() [][]string {
					rows := make([][]string, 0, len(uploaded))
					for _, img := range uploaded {
						rows = append(rows, []string{img.Id, img.Url})
					}
					return rows
				})
		},
	}

	var imageID string
	del := &cobra.Command{
		Use:   "delete",
		Short: "Delete a single image by id",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateImageType(imageType); err != nil {
				return err
			}
			if err := RequireConfirm(confirm, deps.Getenv, "delete image"); err != nil {
				return err
			}
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				return client.DeleteImage(cmd.Context(), app, id, locale, imageType, imageID)
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, map[string]string{"deleted": imageID},
				[]string{"DELETED"}, func() [][]string { return [][]string{{imageID}} })
		},
	}
	del.Flags().StringVar(&imageID, "image-id", "", "id of the image to delete (required)")
	_ = del.MarkFlagRequired("image-id")

	deleteAll := &cobra.Command{
		Use:   "delete-all",
		Short: "Delete every image of a locale and image type",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateImageType(imageType); err != nil {
				return err
			}
			if err := RequireConfirm(confirm, deps.Getenv, "delete all images"); err != nil {
				return err
			}
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var deleted []*androidpublisher.Image
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				var inner error
				deleted, inner = client.DeleteAllImages(cmd.Context(), app, id, locale, imageType)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, deleted,
				[]string{"DELETED_ID"},
				func() [][]string {
					rows := make([][]string, 0, len(deleted))
					for _, img := range deleted {
						rows = append(rows, []string{img.Id})
					}
					return rows
				})
		},
	}

	for _, c := range []*cobra.Command{list, upload, del, deleteAll} {
		c.Flags().StringVar(&app, "app", "", "package name of the app (required)")
		_ = c.MarkFlagRequired("app")
		c.Flags().StringVar(&locale, "locale", "", "BCP-47 locale, e.g. en-US (required)")
		_ = c.MarkFlagRequired("locale")
		c.Flags().StringVar(&imageType, "type", "", "image type, e.g. phoneScreenshots, icon, featureGraphic (required)")
		_ = c.MarkFlagRequired("type")
	}
	for _, c := range []*cobra.Command{upload, del, deleteAll} {
		c.Flags().StringVar(&editID, "edit-id", "", "reuse an existing edit transaction (no auto-commit)")
	}
	for _, c := range []*cobra.Command{del, deleteAll} {
		c.Flags().BoolVar(&confirm, "confirm", false, "confirm the destructive action")
	}

	imagesCmd.AddCommand(list, upload, del, deleteAll)
	return imagesCmd
}
