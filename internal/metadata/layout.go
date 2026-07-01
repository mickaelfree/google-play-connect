// Package metadata maps Play store listings and images to a canonical
// on-disk tree so they can be edited offline and re-applied atomically
// (gpc metadata pull / push / validate).
package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Listing is the offline representation of one locale's store listing.
type Listing struct {
	Title            string
	ShortDescription string
	FullDescription  string
	Video            string
}

// ScreenshotTypes are stored as directories of numbered image files.
var ScreenshotTypes = []string{
	"phoneScreenshots", "sevenInchScreenshots", "tenInchScreenshots",
	"tvScreenshots", "wearScreenshots",
}

// SingleImageTypes are stored as flat <type>.png files.
var SingleImageTypes = []string{"icon", "featureGraphic", "promoGraphic", "tvBanner"}

// AppDir returns the root of one app's metadata tree.
func AppDir(root, packageName string) string {
	return filepath.Join(root, packageName)
}

func listingDir(root, pkg, locale string) string {
	return filepath.Join(AppDir(root, pkg), "listings", locale)
}

// listingFiles maps Listing fields to their file names.
var listingFiles = []struct {
	name string
	get  func(Listing) string
	set  func(*Listing, string)
}{
	{"title.txt", func(l Listing) string { return l.Title }, func(l *Listing, v string) { l.Title = v }},
	{"short_description.txt", func(l Listing) string { return l.ShortDescription }, func(l *Listing, v string) { l.ShortDescription = v }},
	{"full_description.txt", func(l Listing) string { return l.FullDescription }, func(l *Listing, v string) { l.FullDescription = v }},
	{"video.txt", func(l Listing) string { return l.Video }, func(l *Listing, v string) { l.Video = v }},
}

// WriteListing writes one locale's listing files. Empty fields produce no
// file (and any stale file for a now-empty field is removed).
func WriteListing(root, pkg, locale string, l Listing) error {
	dir := listingDir(root, pkg, locale)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	for _, f := range listingFiles {
		path := filepath.Join(dir, f.name)
		value := f.get(l)
		if value == "" {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("remove %s: %w", path, err)
			}
			continue
		}
		if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// ReadListing reads one locale's listing files; missing files read as "".
func ReadListing(root, pkg, locale string) (Listing, error) {
	dir := listingDir(root, pkg, locale)
	var l Listing
	for _, f := range listingFiles {
		data, err := os.ReadFile(filepath.Join(dir, f.name))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return Listing{}, fmt.Errorf("read %s: %w", f.name, err)
		}
		f.set(&l, string(data))
	}
	return l, nil
}

// ListingLocales returns the locales that have a listings directory, sorted.
func ListingLocales(root, pkg string) ([]string, error) {
	return subDirs(filepath.Join(AppDir(root, pkg), "listings"))
}

// ImageLocales returns the locales that have an images directory, sorted.
func ImageLocales(root, pkg string) ([]string, error) {
	return subDirs(filepath.Join(AppDir(root, pkg), "images"))
}

func subDirs(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", dir, err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	slices.Sort(names)
	return names, nil
}

// LocalImages returns the local image files for a locale + type, sorted by
// file name. Screenshot types read a directory; single-image types read a
// flat <type>.png/.jpg file. A missing type returns an empty slice.
func LocalImages(root, pkg, locale, imageType string) ([]string, error) {
	base := filepath.Join(AppDir(root, pkg), "images", locale)
	if slices.Contains(ScreenshotTypes, imageType) {
		dir := filepath.Join(base, imageType)
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, nil
			}
			return nil, fmt.Errorf("read %s: %w", dir, err)
		}
		paths := make([]string, 0, len(entries))
		for _, e := range entries {
			if e.IsDir() || strings.HasSuffix(e.Name(), ".json") {
				continue
			}
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
		slices.Sort(paths)
		return paths, nil
	}
	for _, ext := range []string{".png", ".jpg", ".jpeg"} {
		path := filepath.Join(base, imageType+ext)
		if _, err := os.Stat(path); err == nil {
			return []string{path}, nil
		}
	}
	return nil, nil
}

// ImageManifestEntry records one remote image in a pull manifest.
type ImageManifestEntry struct {
	ID   string `json:"id"`
	Sha1 string `json:"sha1"`
	URL  string `json:"url"`
}

// WriteImageManifest persists the remote image state for a locale + type as
// <type>.manifest.json, giving pull a faithful reference without binaries.
func WriteImageManifest(root, pkg, locale, imageType string, entries []ImageManifestEntry) error {
	dir := filepath.Join(AppDir(root, pkg), "images", locale)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", dir, err)
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	path := filepath.Join(dir, imageType+".manifest.json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
