package cmd

import "fmt"

// RequireConfirm gates destructive actions: the caller must pass --confirm,
// unless running in CI (env CI=true) where prompting is impossible.
func RequireConfirm(confirm bool, getenv func(string) string, action string) error {
	if confirm || getenv("CI") == "true" {
		return nil
	}
	return fmt.Errorf("%s is destructive: re-run with --confirm (or set CI=true in automation)", action)
}
