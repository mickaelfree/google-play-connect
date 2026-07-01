package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

func newBundlesCmd(deps Deps, flags *RootFlags) *cobra.Command {
	bundlesCmd := &cobra.Command{
		Use:   "bundles",
		Short: "Upload and inspect Android App Bundles (.aab)",
	}

	var app, editID, file string

	list := &cobra.Command{
		Use:   "list",
		Short: "List bundles attached to the app",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var bundles []*androidpublisher.Bundle
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(id string) error {
				var inner error
				bundles, inner = client.ListBundles(cmd.Context(), app, id)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, bundles,
				[]string{"VERSION_CODE", "SHA256"},
				func() [][]string {
					rows := make([][]string, 0, len(bundles))
					for _, b := range bundles {
						rows = append(rows, []string{strconv.FormatInt(b.VersionCode, 10), b.Sha256})
					}
					return rows
				})
		},
	}

	upload := &cobra.Command{
		Use:   "upload",
		Short: "Upload an .aab into an edit (auto-commits unless --edit-id)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var bundle *androidpublisher.Bundle
			err = client.WithTransaction(cmd.Context(), app, editID, func(id string) error {
				f, openErr := os.Open(file)
				if openErr != nil {
					return fmt.Errorf("open %s: %w", file, openErr)
				}
				defer f.Close()
				var inner error
				bundle, inner = client.UploadBundle(cmd.Context(), app, id, f)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, bundle,
				[]string{"VERSION_CODE", "SHA256"},
				func() [][]string {
					return [][]string{{strconv.FormatInt(bundle.VersionCode, 10), bundle.Sha256}}
				})
		},
	}
	upload.Flags().StringVar(&file, "file", "", "path to the .aab file (required)")
	_ = upload.MarkFlagRequired("file")
	upload.Flags().StringVar(&editID, "edit-id", "", "reuse an existing edit transaction (no auto-commit)")

	for _, c := range []*cobra.Command{list, upload} {
		c.Flags().StringVar(&app, "app", "", "package name of the app (required)")
		_ = c.MarkFlagRequired("app")
	}

	bundlesCmd.AddCommand(list, upload)
	return bundlesCmd
}
