package cmd

import (
	"github.com/spf13/cobra"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// newAppsCmd exposes app-level information. Note: the Android Publisher API
// has no endpoint to list all apps of an account, so unlike asc there is no
// `apps list` — commands always take an explicit --app package name.
func newAppsCmd(deps Deps, flags *RootFlags) *cobra.Command {
	appsCmd := &cobra.Command{
		Use:   "apps",
		Short: "App-level information (details, default language, contacts)",
	}

	var app string
	details := &cobra.Command{
		Use:   "details",
		Short: "Show AppDetails (default language, contact info)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			var details *androidpublisher.AppDetails
			err = client.WithReadOnlyEdit(cmd.Context(), app, func(editID string) error {
				var inner error
				details, inner = client.GetDetails(cmd.Context(), app, editID)
				return inner
			})
			if err != nil {
				return err
			}
			return renderResult(deps, flags, details,
				[]string{"DEFAULT_LANGUAGE", "CONTACT_EMAIL", "CONTACT_WEBSITE"},
				func() [][]string {
					return [][]string{{details.DefaultLanguage, details.ContactEmail, details.ContactWebsite}}
				})
		},
	}
	details.Flags().StringVar(&app, "app", "", "package name of the app (required)")
	_ = details.MarkFlagRequired("app")

	appsCmd.AddCommand(details)
	return appsCmd
}
