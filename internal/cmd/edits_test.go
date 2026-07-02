package cmd_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/cmd"
	"github.com/mickaelfree/google-play-connect/internal/playapi"
	"github.com/mickaelfree/google-play-connect/internal/playapi/playapitest"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// newTestRoot builds a gpc root command whose Play client talks to a fake
// server driven by handler. Output is captured in the returned buffer.
// Later command-group tests (listings, images, tracks...) reuse this helper.
func newTestRoot(t *testing.T, handler http.Handler) (root interface {
	SetArgs([]string)
	Execute() error
}, buf *bytes.Buffer) {
	t.Helper()
	svc, _ := playapitest.NewService(t, handler)
	out := &bytes.Buffer{}
	deps := cmd.Deps{
		Stdout: out,
		Getenv: func(string) string { return "" },
		NewClient: func(ctx context.Context, serviceAccountPath string) (*playapi.Client, error) {
			return playapi.NewClient(svc), nil
		},
	}
	return cmd.NewRootCmd(deps), out
}

func TestEditsBeginOutputsJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "edit-9", ExpiryTimeSeconds: "3600"})
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"edits", "begin", "--app", "com.example.app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	var got androidpublisher.AppEdit
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("output is not JSON: %v\n%s", err, out.String())
	}
	if got.Id != "edit-9" {
		t.Fatalf("got id %q, want edit-9", got.Id)
	}
}

func TestEditsDiscardRequiresConfirm(t *testing.T) {
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{"edits", "discard", "--app", "com.example.app", "--edit-id", "edit-9"})
	if err := root.Execute(); err == nil {
		t.Fatal("discard without --confirm must fail")
	}
}

func TestEditsCommit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/edit-9:commit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "edit-9"})
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"edits", "commit", "--app", "com.example.app", "--edit-id", "edit-9"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("edit-9")) {
		t.Fatalf("output missing edit id: %s", out.String())
	}
}

func TestEditsGet(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/edit-9", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "edit-9", ExpiryTimeSeconds: "3600"})
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"edits", "get", "--app", "com.example.app", "--edit-id", "edit-9"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("edit-9")) {
		t.Fatalf("output missing edit id: %s", out.String())
	}
}

func TestAppsDetailsUsesReadOnlyEdit(t *testing.T) {
	discarded := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tmp-edit"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp-edit/details", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppDetails{DefaultLanguage: "fr-FR", ContactEmail: "dev@example.com"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp-edit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			discarded = true
		}
		w.WriteHeader(http.StatusOK)
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"apps", "details", "--app", "com.example.app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("fr-FR")) {
		t.Fatalf("output missing details: %s", out.String())
	}
	if !discarded {
		t.Fatal("read-only edit was not discarded")
	}
}
