package metadata_test

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mickaelfree/google-play-connect/internal/metadata"
)

// writePNG creates a real PNG of the given size at path.
func writePNG(t *testing.T, path string, w, h int) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
}

func TestValidateListingLimits(t *testing.T) {
	bad := metadata.Listing{
		Title:            strings.Repeat("x", metadata.MaxTitleLen+1),
		ShortDescription: strings.Repeat("y", metadata.MaxShortDescriptionLen+1),
		FullDescription:  strings.Repeat("z", metadata.MaxFullDescriptionLen+1),
	}
	issues := metadata.ValidateListing("fr-FR", bad)
	if len(issues) != 3 {
		t.Fatalf("want 3 issues, got %d: %+v", len(issues), issues)
	}
	for _, issue := range issues {
		if issue.Locale != "fr-FR" {
			t.Fatalf("issue missing locale: %+v", issue)
		}
	}
}

func TestValidateListingEmptyTitle(t *testing.T) {
	issues := metadata.ValidateListing("en-US", metadata.Listing{Title: ""})
	if len(issues) != 1 || issues[0].Field != "title" {
		t.Fatalf("empty title must be flagged: %+v", issues)
	}
}

func TestValidateListingOK(t *testing.T) {
	ok := metadata.Listing{Title: "My App", ShortDescription: "Nice", FullDescription: "Long text"}
	if issues := metadata.ValidateListing("en-US", ok); len(issues) != 0 {
		t.Fatalf("valid listing flagged: %+v", issues)
	}
}

func TestValidateTree(t *testing.T) {
	root := t.TempDir()
	if err := metadata.WriteListing(root, "com.example.app", "en-US", metadata.Listing{Title: strings.Repeat("x", 31)}); err != nil {
		t.Fatal(err)
	}
	if err := metadata.WriteListing(root, "com.example.app", "fr-FR", metadata.Listing{Title: "OK"}); err != nil {
		t.Fatal(err)
	}
	issues, err := metadata.ValidateTree(root, "com.example.app")
	if err != nil {
		t.Fatalf("ValidateTree: %v", err)
	}
	if len(issues) != 1 || issues[0].Locale != "en-US" {
		t.Fatalf("want exactly the en-US title issue, got %+v", issues)
	}
}

func TestValidateImagesDimensions(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(metadata.AppDir(root, "com.example.app"), "images", "en-US")
	writePNG(t, filepath.Join(base, "icon.png"), 100, 100)                           // wrong: must be 512x512
	writePNG(t, filepath.Join(base, "featureGraphic.png"), 1024, 500)                // correct
	writePNG(t, filepath.Join(base, "phoneScreenshots", "1.png"), 200, 400)          // wrong: side < 320
	writePNG(t, filepath.Join(base, "phoneScreenshots", "2.png"), 1080, 1920)        // correct

	issues, err := metadata.ValidateImages(root, "com.example.app")
	if err != nil {
		t.Fatalf("ValidateImages: %v", err)
	}
	if len(issues) != 2 {
		t.Fatalf("want 2 issues (icon size, screenshot 1 too small), got %+v", issues)
	}
	for _, issue := range issues {
		if issue.Locale != "en-US" {
			t.Fatalf("issue missing locale: %+v", issue)
		}
	}
}

func TestValidateImagesRejectsBadFormat(t *testing.T) {
	root := t.TempDir()
	base := filepath.Join(metadata.AppDir(root, "com.example.app"), "images", "en-US")
	path := filepath.Join(base, "phoneScreenshots", "1.png")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("this is not an image"), 0o644); err != nil {
		t.Fatal(err)
	}
	issues, err := metadata.ValidateImages(root, "com.example.app")
	if err != nil {
		t.Fatalf("ValidateImages: %v", err)
	}
	if len(issues) != 1 || issues[0].Field != "phoneScreenshots" {
		t.Fatalf("undecodable file must be flagged, got %+v", issues)
	}
}

func TestValidateTreeIncludesImages(t *testing.T) {
	root := t.TempDir()
	if err := metadata.WriteListing(root, "com.example.app", "en-US", metadata.Listing{Title: "OK"}); err != nil {
		t.Fatal(err)
	}
	base := filepath.Join(metadata.AppDir(root, "com.example.app"), "images", "en-US")
	writePNG(t, filepath.Join(base, "icon.png"), 100, 100)
	issues, err := metadata.ValidateTree(root, "com.example.app")
	if err != nil {
		t.Fatalf("ValidateTree: %v", err)
	}
	if len(issues) != 1 || issues[0].Field != "icon" {
		t.Fatalf("ValidateTree must include image issues, got %+v", issues)
	}
}
