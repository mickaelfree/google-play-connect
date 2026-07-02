package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func TestTracksList(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tmp"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp/tracks", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.TracksListResponse{
			Tracks: []*androidpublisher.Track{{Track: "production"}, {Track: "beta"}},
		})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"tracks", "list", "--app", "com.example.app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("production")) {
		t.Fatalf("output missing tracks: %s", out.String())
	}
}

func TestTracksUpdateRollout(t *testing.T) {
	committed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/tracks/production", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(androidpublisher.Track{
				Track: "production",
				Releases: []*androidpublisher.TrackRelease{
					{Status: "inProgress", UserFraction: 0.1, VersionCodes: googleapi.Int64s{42}},
				},
			})
			return
		}
		var body androidpublisher.Track
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		rel := body.Releases[0]
		if rel.UserFraction != 0.5 || rel.Status != "inProgress" || rel.VersionCodes[0] != 42 {
			t.Errorf("patched release wrong: %+v", rel)
		}
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		committed = true
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{
		"tracks", "update",
		"--app", "com.example.app",
		"--track", "production",
		"--rollout", "0.5",
		"--confirm",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !committed {
		t.Fatal("tracks update must commit")
	}
}

func TestTracksUpdateRequiresConfirm(t *testing.T) {
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{"tracks", "update", "--app", "com.example.app", "--track", "production", "--rollout", "0.5"})
	if err := root.Execute(); err == nil {
		t.Fatal("tracks update without --confirm must fail")
	}
}

func TestReleasesPublishRequiresConfirm(t *testing.T) {
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{
		"releases", "publish",
		"--app", "com.example.app",
		"--track", "production",
		"--version-codes", "42",
	})
	if err := root.Execute(); err == nil {
		t.Fatal("publish without --confirm must fail")
	}
}

func TestReleasesPublishFullFlow(t *testing.T) {
	notesDir := t.TempDir()
	notesPath := filepath.Join(notesDir, "en-US.txt")
	if err := os.WriteFile(notesPath, []byte("Bug fixes and improvements"), 0o644); err != nil {
		t.Fatal(err)
	}

	committed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/tracks/production", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method %s", r.Method)
		}
		var body androidpublisher.Track
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		rel := body.Releases[0]
		if rel.Status != "completed" || rel.VersionCodes[0] != 42 {
			t.Errorf("unexpected release: %+v", rel)
		}
		if len(rel.ReleaseNotes) != 1 || rel.ReleaseNotes[0].Text != "Bug fixes and improvements" {
			t.Errorf("unexpected notes: %+v", rel.ReleaseNotes)
		}
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		committed = true
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{
		"releases", "publish",
		"--app", "com.example.app",
		"--track", "production",
		"--version-codes", "42",
		"--notes-file", "en-US=" + notesPath,
		"--confirm",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !committed {
		t.Fatal("publish must commit")
	}
}

func TestReleasesPublishRolloutSetsInProgress(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/tracks/production", func(w http.ResponseWriter, r *http.Request) {
		var body androidpublisher.Track
		_ = json.NewDecoder(r.Body).Decode(&body)
		rel := body.Releases[0]
		if rel.Status != "inProgress" || rel.UserFraction != 0.25 {
			t.Errorf("rollout release wrong: %+v", rel)
		}
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{
		"releases", "publish",
		"--app", "com.example.app",
		"--track", "production",
		"--version-codes", "42",
		"--rollout", "0.25",
		"--confirm",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

func TestReleasesPublishNotesDir(t *testing.T) {
	notesDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(notesDir, "en-US.txt"), []byte("EN notes"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(notesDir, "fr-FR.txt"), []byte("Notes FR"), 0o644); err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/tracks/beta", func(w http.ResponseWriter, r *http.Request) {
		var body androidpublisher.Track
		_ = json.NewDecoder(r.Body).Decode(&body)
		notes := body.Releases[0].ReleaseNotes
		if len(notes) != 2 {
			t.Errorf("want 2 locales of notes, got %+v", notes)
		}
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{
		"releases", "publish",
		"--app", "com.example.app",
		"--track", "beta",
		"--version-codes", "42",
		"--notes-dir", notesDir,
		"--confirm",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
}

func TestReleasesPublishExplicitZeroRolloutRejected(t *testing.T) {
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{
		"releases", "publish",
		"--app", "com.example.app",
		"--track", "production",
		"--version-codes", "42",
		"--rollout", "0",
		"--confirm",
	})
	if err := root.Execute(); err == nil {
		t.Fatal("explicit --rollout 0 must be rejected, not silently become a full rollout")
	}
}
