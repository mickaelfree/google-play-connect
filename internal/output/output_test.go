package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/output"
)

func TestResolveExplicitFlag(t *testing.T) {
	f, err := output.Resolve("json", true)
	if err != nil || f != output.FormatJSON {
		t.Fatalf("got %v/%v, want json", f, err)
	}
	f, err = output.Resolve("table", false)
	if err != nil || f != output.FormatTable {
		t.Fatalf("got %v/%v, want table", f, err)
	}
}

func TestResolveAutoDetect(t *testing.T) {
	f, err := output.Resolve("", true)
	if err != nil || f != output.FormatTable {
		t.Fatalf("TTY should default to table, got %v/%v", f, err)
	}
	f, err = output.Resolve("", false)
	if err != nil || f != output.FormatJSON {
		t.Fatalf("non-TTY should default to json, got %v/%v", f, err)
	}
}

func TestResolveInvalid(t *testing.T) {
	if _, err := output.Resolve("yaml", false); err == nil {
		t.Fatal("expected error for invalid format")
	}
}

func TestRenderJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := output.RenderJSON(&buf, map[string]string{"id": "edit-1"}); err != nil {
		t.Fatalf("RenderJSON: %v", err)
	}
	if !strings.Contains(buf.String(), `"id": "edit-1"`) {
		t.Fatalf("unexpected JSON: %s", buf.String())
	}
	if !strings.HasSuffix(buf.String(), "\n") {
		t.Fatal("JSON output must end with a newline")
	}
}

func TestRenderTable(t *testing.T) {
	var buf bytes.Buffer
	err := output.RenderTable(&buf, []string{"LOCALE", "TITLE"}, [][]string{
		{"en-US", "My App"},
		{"fr-FR", "Mon App"},
	})
	if err != nil {
		t.Fatalf("RenderTable: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "LOCALE") || !strings.Contains(got, "fr-FR") {
		t.Fatalf("unexpected table: %s", got)
	}
}
