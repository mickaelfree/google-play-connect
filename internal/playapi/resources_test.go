package playapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/playapi"
	"github.com/mickaelfree/google-play-connect/internal/playapi/playapitest"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
)

func TestUploadImage(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload/androidpublisher/v3/applications/com.example.app/edits/edit-1/listings/en-US/phoneScreenshots", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		// The SDK sends uploadType=multipart: a multipart/related body with a
		// JSON metadata part followed by the media part. Assert the media
		// bytes are present rather than parsing the multipart envelope.
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}
		if !bytes.Contains(body, []byte("fake-png-bytes")) {
			t.Errorf("media bytes missing from multipart body: %q", body)
		}
		json.NewEncoder(w).Encode(androidpublisher.ImagesUploadResponse{
			Image: &androidpublisher.Image{Id: "img-1", Url: "https://example.com/img-1.png"},
		})
	})

	svc, _ := playapitest.NewService(t, mux)
	client := playapi.NewClient(svc)

	img, err := client.UploadImage(context.Background(), "com.example.app", "edit-1", "en-US", playapi.ImageTypePhoneScreenshots, bytes.NewReader([]byte("fake-png-bytes")))
	if err != nil {
		t.Fatalf("UploadImage: %v", err)
	}
	if img.Id != "img-1" {
		t.Fatalf("got image id %q, want img-1", img.Id)
	}
}

func TestUpdateTrackRoundTrip(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/edit-1/tracks/production", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("unexpected method %s", r.Method)
		}
		var body androidpublisher.Track
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
			return
		}
		if len(body.Releases) != 1 || body.Releases[0].VersionCodes[0] != 42 {
			t.Errorf("unexpected releases in request body: %+v", body.Releases)
		}
		json.NewEncoder(w).Encode(body)
	})

	svc, _ := playapitest.NewService(t, mux)
	client := playapi.NewClient(svc)

	track := &androidpublisher.Track{
		Track: playapi.TrackProduction,
		Releases: []*androidpublisher.TrackRelease{
			{
				Status:       playapi.ReleaseStatusCompleted,
				VersionCodes: googleapi.Int64s{42},
				ReleaseNotes: playapi.UpsertReleaseNotes(nil, "en-US", "Bug fixes"),
			},
		},
	}

	updated, err := client.UpdateTrack(context.Background(), "com.example.app", "edit-1", "production", track)
	if err != nil {
		t.Fatalf("UpdateTrack: %v", err)
	}
	if updated.Releases[0].ReleaseNotes[0].Text != "Bug fixes" {
		t.Fatalf("unexpected release notes: %+v", updated.Releases[0].ReleaseNotes)
	}
}

func TestUploadBundle(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload/androidpublisher/v3/applications/com.example.app/edits/edit-1/bundles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method %s", r.Method)
		}
		json.NewEncoder(w).Encode(androidpublisher.Bundle{VersionCode: 7, Sha256: "deadbeef"})
	})

	svc, _ := playapitest.NewService(t, mux)
	client := playapi.NewClient(svc)

	bundle, err := client.UploadBundle(context.Background(), "com.example.app", "edit-1", bytes.NewReader([]byte("fake-aab-bytes")))
	if err != nil {
		t.Fatalf("UploadBundle: %v", err)
	}
	if bundle.VersionCode != 7 {
		t.Fatalf("got version code %d, want 7", bundle.VersionCode)
	}
}

func TestListListings(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/edits/edit-1/listings", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.ListingsListResponse{
			Listings: []*androidpublisher.Listing{
				{Language: "en-US", Title: "My App"},
				{Language: "fr-FR", Title: "Mon App"},
			},
		})
	})

	svc, _ := playapitest.NewService(t, mux)
	client := playapi.NewClient(svc)

	listings, err := client.ListListings(context.Background(), "com.example.app", "edit-1")
	if err != nil {
		t.Fatalf("ListListings: %v", err)
	}
	if len(listings) != 2 || listings[1].Title != "Mon App" {
		t.Fatalf("unexpected listings: %+v", listings)
	}
}

func TestListReleaseSummaries(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/androidpublisher/v3/applications/com.example.app/tracks/production/releases", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(androidpublisher.ListReleaseSummariesResponse{
			Releases: []*androidpublisher.ReleaseSummary{
				{ReleaseName: "42 (1.2.3)", ReleaseLifecycleState: "ACTIVE", Track: "production"},
			},
		})
	})

	svc, _ := playapitest.NewService(t, mux)
	client := playapi.NewClient(svc)

	releases, err := client.ListReleaseSummaries(context.Background(), "com.example.app", "production")
	if err != nil {
		t.Fatalf("ListReleaseSummaries: %v", err)
	}
	if len(releases) != 1 || releases[0].ReleaseName != "42 (1.2.3)" {
		t.Fatalf("unexpected releases: %+v", releases)
	}
}

func TestUpsertReleaseNotesDoesNotMutate(t *testing.T) {
	original := []*androidpublisher.LocalizedText{{Language: "en-US", Text: "old"}}
	updated := playapi.UpsertReleaseNotes(original, "en-US", "new")
	if original[0].Text != "old" {
		t.Fatal("UpsertReleaseNotes mutated its input")
	}
	if updated[0].Text != "new" {
		t.Fatalf("got %q, want new", updated[0].Text)
	}
	added := playapi.UpsertReleaseNotes(original, "fr-FR", "nouveau")
	if len(added) != 2 || added[1].Language != "fr-FR" {
		t.Fatalf("unexpected append result: %+v", added)
	}
}
