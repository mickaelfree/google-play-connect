package playapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/playapi"
	"github.com/mickaelfree/google-play-connect/internal/playapi/playapitest"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

func TestBeginCommitDiscardEdit(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "edit-1", ExpiryTimeSeconds: "3600"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/edit-1:commit", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "edit-1"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/edit-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})

	svc, _ := playapitest.NewService(t, mux)
	client := playapi.NewClient(svc)
	ctx := context.Background()

	edit, err := client.BeginEdit(ctx, "com.example.app")
	if err != nil {
		t.Fatalf("BeginEdit: %v", err)
	}
	if edit.Id != "edit-1" {
		t.Fatalf("got edit id %q, want edit-1", edit.Id)
	}

	if _, err := client.CommitEdit(ctx, "com.example.app", edit.Id); err != nil {
		t.Fatalf("CommitEdit: %v", err)
	}

	if err := client.DiscardEdit(ctx, "com.example.app", edit.Id); err != nil {
		t.Fatalf("DiscardEdit: %v", err)
	}
}

func TestWithTransactionSharedEditID(t *testing.T) {
	svc, _ := playapitest.NewService(t, http.NewServeMux())
	client := playapi.NewClient(svc)

	called := false
	err := client.WithTransaction(context.Background(), "com.example.app", "existing-edit", func(editID string) error {
		called = true
		if editID != "existing-edit" {
			t.Fatalf("got editID %q, want existing-edit", editID)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithTransaction: %v", err)
	}
	if !called {
		t.Fatal("fn was not called")
	}
}
