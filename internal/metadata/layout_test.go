package metadata_test

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/metadata"
)

func TestListingRoundTrip(t *testing.T) {
	root := t.TempDir()
	in := metadata.Listing{
		Title:            "Mon App",
		ShortDescription: "Courte description",
		FullDescription:  "Longue description\nsur deux lignes",
		Video:            "https://youtu.be/xyz",
	}
	if err := metadata.WriteListing(root, "com.example.app", "fr-FR", in); err != nil {
		t.Fatalf("WriteListing: %v", err)
	}
	out, err := metadata.ReadListing(root, "com.example.app", "fr-FR")
	if err != nil {
		t.Fatalf("ReadListing: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("round trip mismatch:\n in=%+v\nout=%+v", in, out)
	}
}

func TestReadListingMissingVideoIsEmpty(t *testing.T) {
	root := t.TempDir()
	if err := metadata.WriteListing(root, "com.example.app", "en-US", metadata.Listing{Title: "X"}); err != nil {
		t.Fatal(err)
	}
	// video.txt for an empty video must not exist, and reads as "".
	videoPath := filepath.Join(metadata.AppDir(root, "com.example.app"), "listings", "en-US", "video.txt")
	if _, err := os.Stat(videoPath); !os.IsNotExist(err) {
		t.Fatalf("empty video should not create video.txt: %v", err)
	}
	out, err := metadata.ReadListing(root, "com.example.app", "en-US")
	if err != nil || out.Video != "" {
		t.Fatalf("got %+v / %v", out, err)
	}
}

func TestListingLocalesSorted(t *testing.T) {
	root := t.TempDir()
	for _, loc := range []string{"fr-FR", "de-DE", "en-US"} {
		if err := metadata.WriteListing(root, "com.example.app", loc, metadata.Listing{Title: loc}); err != nil {
			t.Fatal(err)
		}
	}
	locales, err := metadata.ListingLocales(root, "com.example.app")
	if err != nil {
		t.Fatalf("ListingLocales: %v", err)
	}
	want := []string{"de-DE", "en-US", "fr-FR"}
	if !reflect.DeepEqual(locales, want) {
		t.Fatalf("got %v, want %v", locales, want)
	}
}

func TestLocalImagesScreenshotDirAndSingleFile(t *testing.T) {
	root := t.TempDir()
	shotDir := filepath.Join(metadata.AppDir(root, "com.example.app"), "images", "en-US", "phoneScreenshots")
	if err := os.MkdirAll(shotDir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"2.png", "1.png"} {
		if err := os.WriteFile(filepath.Join(shotDir, name), []byte("png"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	iconPath := filepath.Join(metadata.AppDir(root, "com.example.app"), "images", "en-US", "icon.png")
	if err := os.WriteFile(iconPath, []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}

	shots, err := metadata.LocalImages(root, "com.example.app", "en-US", "phoneScreenshots")
	if err != nil {
		t.Fatalf("LocalImages screenshots: %v", err)
	}
	if len(shots) != 2 || filepath.Base(shots[0]) != "1.png" {
		t.Fatalf("screenshots not sorted: %v", shots)
	}

	icons, err := metadata.LocalImages(root, "com.example.app", "en-US", "icon")
	if err != nil {
		t.Fatalf("LocalImages icon: %v", err)
	}
	if len(icons) != 1 || filepath.Base(icons[0]) != "icon.png" {
		t.Fatalf("unexpected icon result: %v", icons)
	}

	none, err := metadata.LocalImages(root, "com.example.app", "en-US", "featureGraphic")
	if err != nil || len(none) != 0 {
		t.Fatalf("missing type should be empty, got %v / %v", none, err)
	}
}
