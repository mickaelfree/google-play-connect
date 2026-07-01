package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/mickaelfree/google-play-connect/skillsfs"
)

func newInstallSkillsCmd(deps Deps, flags *RootFlags) *cobra.Command {
	var dir string
	installCmd := &cobra.Command{
		Use:   "install-skills",
		Short: "Install gpc's AI agent skills into a skills directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			var installed []string
			err := fs.WalkDir(skillsfs.FS, "skills", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				rel, relErr := filepath.Rel("skills", path)
				if relErr != nil {
					return relErr
				}
				target := filepath.Join(dir, rel)
				if mkErr := os.MkdirAll(filepath.Dir(target), 0o755); mkErr != nil {
					return mkErr
				}
				data, readErr := fs.ReadFile(skillsfs.FS, path)
				if readErr != nil {
					return readErr
				}
				if writeErr := os.WriteFile(target, data, 0o644); writeErr != nil {
					return writeErr
				}
				installed = append(installed, filepath.Dir(rel))
				return nil
			})
			if err != nil {
				return fmt.Errorf("install skills: %w", err)
			}
			return renderResult(deps, flags, map[string]any{"installed": installed, "dir": dir},
				[]string{"SKILL"},
				func() [][]string {
					rows := make([][]string, 0, len(installed))
					for _, s := range installed {
						rows = append(rows, []string{s})
					}
					return rows
				})
		},
	}
	home, _ := os.UserHomeDir()
	installCmd.Flags().StringVar(&dir, "dir", filepath.Join(home, ".claude", "skills"), "skills directory to install into")
	return installCmd
}
