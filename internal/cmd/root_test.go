package cmd_test

import (
	"bytes"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/cmd"
)

func testDeps(buf *bytes.Buffer) cmd.Deps {
	return cmd.Deps{
		Stdout: buf,
		Getenv: func(string) string { return "" },
	}
}

func TestRootCmdHasUseGpc(t *testing.T) {
	var buf bytes.Buffer
	root := cmd.NewRootCmd(testDeps(&buf))
	if root.Use != "gpc" {
		t.Fatalf("got Use %q, want gpc", root.Use)
	}
}

func TestRootCmdHelpRuns(t *testing.T) {
	var buf bytes.Buffer
	root := cmd.NewRootCmd(testDeps(&buf))
	root.SetOut(&buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute --help: %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("Google Play")) {
		t.Fatalf("help output missing description: %s", buf.String())
	}
}
