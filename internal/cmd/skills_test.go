package cmd_test

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestInstallSkillsWritesSkillDirs(t *testing.T) {
	target := t.TempDir()
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{"install-skills", "--dir", target})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	for _, name := range []string{"gpc-cli-usage", "gpc-auth-setup", "gpc-metadata-sync", "gpc-release-flow"} {
		path := filepath.Join(target, name, "SKILL.md")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("skill %s not installed: %v", name, err)
		}
		if len(data) == 0 {
			t.Fatalf("skill %s is empty", name)
		}
	}
}
