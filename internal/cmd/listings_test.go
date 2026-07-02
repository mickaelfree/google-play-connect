package cmd_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

func TestListingsListReadOnly(t *testing.T) {
	discarded := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tmp"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp/listings", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.ListingsListResponse{
			Listings: []*androidpublisher.Listing{{Language: "en-US", Title: "My App"}},
		})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			discarded = true
		}
		w.WriteHeader(http.StatusOK)
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"listings", "list", "--app", "com.example.app"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("My App")) {
		t.Fatalf("output missing listing: %s", out.String())
	}
	if !discarded {
		t.Fatal("read path must discard its edit")
	}
}

func TestListingsGet(t *testing.T) {
	discarded := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tmp"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp/listings/fr-FR", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.Listing{Language: "fr-FR", Title: "Mon App"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tmp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			discarded = true
		}
		w.WriteHeader(http.StatusOK)
	})

	root, out := newTestRoot(t, mux)
	root.SetArgs([]string{"listings", "get", "--app", "com.example.app", "--locale", "fr-FR"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !bytes.Contains(out.Bytes(), []byte("Mon App")) {
		t.Fatalf("output missing title: %s", out.String())
	}
	if !discarded {
		t.Fatal("read path must discard its edit")
	}
}

func TestListingsDeleteRequiresConfirmAndDeletes(t *testing.T) {
	root, _ := newTestRoot(t, http.NewServeMux())
	root.SetArgs([]string{"listings", "delete", "--app", "com.example.app", "--locale", "fr-FR"})
	if err := root.Execute(); err == nil {
		t.Fatal("delete without --confirm must fail")
	}

	deleted := false
	committed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/listings/fr-FR", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method %s", r.Method)
		}
		deleted = true
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		committed = true
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root2, _ := newTestRoot(t, mux)
	root2.SetArgs([]string{"listings", "delete", "--app", "com.example.app", "--locale", "fr-FR", "--confirm"})
	if err := root2.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !deleted {
		t.Fatal("listing was not deleted")
	}
	if !committed {
		t.Fatal("delete must commit its transaction")
	}
}

func TestListingsUpdateCommitsTransaction(t *testing.T) {
	committed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx/listings/fr-FR", func(w http.ResponseWriter, r *http.Request) {
		// The update command pre-fetches the listing (GET, empty body) before
		// PUTting the merged result; simulate a brand-new locale with a 404.
		if r.Method == http.MethodGet {
			http.Error(w, `{"error":{"code":404,"message":"not found"}}`, http.StatusNotFound)
			return
		}
		var body androidpublisher.Listing
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
			return
		}
		if body.Title != "Mon App" || body.ShortDescription != "Courte" {
			t.Errorf("unexpected listing body: %+v", body)
		}
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/tx:commit", func(w http.ResponseWriter, r *http.Request) {
		committed = true
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "tx"})
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{
		"listings", "update",
		"--app", "com.example.app",
		"--locale", "fr-FR",
		"--title", "Mon App",
		"--short-description", "Courte",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !committed {
		t.Fatal("update must commit its transaction")
	}
}

func TestListingsUpdateSharedEditDoesNotCommit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/shared/listings/fr-FR", func(w http.ResponseWriter, r *http.Request) {
		var body androidpublisher.Listing
		_ = json.NewDecoder(r.Body).Decode(&body)
		json.NewEncoder(w).Encode(body)
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/shared:commit", func(w http.ResponseWriter, r *http.Request) {
		t.Error("must not commit a shared --edit-id transaction")
	})

	root, _ := newTestRoot(t, mux)
	root.SetArgs([]string{
		"listings", "update",
		"--app", "com.example.app",
		"--edit-id", "shared",
		"--locale", "fr-FR",
		"--title", "Mon App",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
}
