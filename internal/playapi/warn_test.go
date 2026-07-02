package playapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/playapi"
	"github.com/mickaelfree/google-play-connect/internal/playapi/playapitest"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

func TestDiscardFailureWarns(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		json.NewEncoder(w).Encode(androidpublisher.AppEdit{Id: "edit-1"})
	})
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/edit-1", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("unexpected method %s", r.Method)
		}
		http.Error(w, `{"error":{"code":500,"message":"boom"}}`, http.StatusInternalServerError)
	})

	svc, _ := playapitest.NewService(t, mux)
	client := playapi.NewClient(svc)

	original := playapi.WarnWriter
	buf := &bytes.Buffer{}
	playapi.WarnWriter = buf
	t.Cleanup(func() { playapi.WarnWriter = original })

	fnErr := errors.New("callback failed")
	err := client.WithTransaction(context.Background(), "com.example.app", "", func(editID string) error {
		return fnErr
	})
	if !errors.Is(err, fnErr) {
		t.Fatalf("got error %v, want fn's error %v", err, fnErr)
	}
	if !bytes.Contains(buf.Bytes(), []byte("failed to discard edit")) {
		t.Fatalf("warning buffer missing discard failure message: %s", buf.String())
	}
}
