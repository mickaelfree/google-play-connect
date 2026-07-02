package cmd

import (
	"github.com/spf13/cobra"

	"github.com/mickaelfree/google-play-connect/internal/output"
)

// renderResult renders v according to the resolved output format. rows is
// only invoked for table output, so JSON stays a faithful dump of v.
func renderResult(deps Deps, flags *RootFlags, v any, headers []string, rows func() [][]string) error {
	format, err := output.Resolve(flags.Output, output.IsTTY(deps.Stdout))
	if err != nil {
		return err
	}
	if format == output.FormatTable && rows != nil {
		return output.RenderTable(deps.Stdout, headers, rows())
	}
	return output.RenderJSON(deps.Stdout, v)
}

func newEditsCmd(deps Deps, flags *RootFlags) *cobra.Command {
	editsCmd := &cobra.Command{
		Use:   "edits",
		Short: "Low-level edit transaction control (begin, commit, discard, get)",
	}

	var app, editID string
	var confirm bool

	begin := &cobra.Command{
		Use:   "begin",
		Short: "Open a new edit transaction and print its id",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			edit, err := client.BeginEdit(cmd.Context(), app)
			if err != nil {
				return err
			}
			return renderResult(deps, flags, edit, []string{"EDIT_ID", "EXPIRY_EPOCH_S"}, func() [][]string {
				return [][]string{{edit.Id, edit.ExpiryTimeSeconds}}
			})
		},
	}

	commit := &cobra.Command{
		Use:   "commit",
		Short: "Commit an edit, making its changes live",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			edit, err := client.CommitEdit(cmd.Context(), app, editID)
			if err != nil {
				return err
			}
			return renderResult(deps, flags, edit, []string{"EDIT_ID"}, func() [][]string {
				return [][]string{{edit.Id}}
			})
		},
	}

	discard := &cobra.Command{
		Use:   "discard",
		Short: "Delete an edit without applying its changes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := RequireConfirm(confirm, deps.Getenv, "discard edit"); err != nil {
				return err
			}
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			if err := client.DiscardEdit(cmd.Context(), app, editID); err != nil {
				return err
			}
			return renderResult(deps, flags, map[string]string{"discarded": editID}, []string{"DISCARDED"}, func() [][]string {
				return [][]string{{editID}}
			})
		},
	}

	get := &cobra.Command{
		Use:   "get",
		Short: "Show an edit's id and expiry",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := deps.NewClient(cmd.Context(), flags.ServiceAccount)
			if err != nil {
				return err
			}
			edit, err := client.GetEdit(cmd.Context(), app, editID)
			if err != nil {
				return err
			}
			return renderResult(deps, flags, edit, []string{"EDIT_ID", "EXPIRY_EPOCH_S"}, func() [][]string {
				return [][]string{{edit.Id, edit.ExpiryTimeSeconds}}
			})
		},
	}

	for _, c := range []*cobra.Command{begin, commit, discard, get} {
		c.Flags().StringVar(&app, "app", "", "package name of the app (required)")
		_ = c.MarkFlagRequired("app")
	}
	for _, c := range []*cobra.Command{commit, discard, get} {
		c.Flags().StringVar(&editID, "edit-id", "", "identifier of the edit (required)")
		_ = c.MarkFlagRequired("edit-id")
	}
	discard.Flags().BoolVar(&confirm, "confirm", false, "confirm the destructive action")

	editsCmd.AddCommand(begin, commit, discard, get)
	return editsCmd
}
