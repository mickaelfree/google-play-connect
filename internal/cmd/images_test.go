package cmd_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

func TestImagesUploadFromFile(t *testing.T) {
	dir := t.TempDir()
	pngPath := filepath.Join(dir, "shot.png")
	if err := os.WriteFile(pngPath, []byte("png-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}

	committed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/upload/androidpublisher/v3/applications/com.example.app/edits/tx/listings/en-US/phoneScreenshots", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte("png-bytes")) {
			t.Errorf("media bytes missing: %q", body)
		}
		json.NewEncoder(w).Encode(androidpublisher.ImagesUploadResponse{
			Image: &androidpublisher.Image{Id: "img-1"},
		})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		committed = true
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{
		"images", "upload",
		"--app", "com.example.app",
		"--locale", "en-US",
		"--type", "phoneScreenshots",
		pngPath,
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("img-1")) {
		t.Fatalf("output missing image id: %s", out.String())
	}
	if !committed {
		t.Fatal("upload must commit its transaction")
	}
}

func TestImagesUploadRejectsBadType(t *testing.T) {
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{
		"images", "upload",
		"--app", "com.example.app",
		"--locale", "en-US",
		"--type", "notAType",
		"whatever.png",
	})
	if err := root.Execute(); err == nil {
		t.Fatal("invalid --type must fail before any API call")
	}
}

func TestImagesDeleteAllRequiresConfirm(t *testing.T) {
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{
		"images", "delete-all",
		"--app", "com.example.app",
		"--locale", "en-US",
		"--type", "phoneScreenshots",
	})
	if err := root.Execute(); err == nil {
		t.Fatal("delete-all without --confirm must fail")
	}
}
