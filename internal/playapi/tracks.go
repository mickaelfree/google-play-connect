package playapi

import (
	"context"
	"fmt"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// Standard track names. Custom and form-factor tracks (e.g. "wear:production")
// are also valid: the API treats track as a plain string.
const (
	TrackInternal   = "internal"
	TrackAlpha      = "alpha"
	TrackBeta       = "beta"
	TrackProduction = "production"
)

// Release status values, as documented on androidpublisher.TrackRelease.Status.
const (
	ReleaseStatusDraft      = "draft"
	ReleaseStatusInProgress = "inProgress"
	ReleaseStatusHalted     = "halted"
	ReleaseStatusCompleted  = "completed"
)

// ListTracks returns every track (including empty ones) in the edit.
func (c *Client) ListTracks(ctx context.Context, packageName, editID string) ([]*androidpublisher.Track, error) {
	resp, err := c.svc.Edits.Tracks.List(packageName, editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list tracks for %s edit %s: %w", packageName, editID, err)
	}
	return resp.Tracks, nil
}

// GetTrack returns a single track with its releases.
func (c *Client) GetTrack(ctx context.Context, packageName, editID, track string) (*androidpublisher.Track, error) {
	t, err := c.svc.Edits.Tracks.Get(packageName, editID, track).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get track %s for %s edit %s: %w", track, packageName, editID, err)
	}
	return t, nil
}

// UpdateTrack replaces a track's releases (rollout, version codes, notes).
func (c *Client) UpdateTrack(ctx context.Context, packageName, editID, track string, t *androidpublisher.Track) (*androidpublisher.Track, error) {
	updated, err := c.svc.Edits.Tracks.Update(packageName, editID, track, t).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("update track %s for %s edit %s: %w", track, packageName, editID, err)
	}
	return updated, nil
}

// UpsertReleaseNotes returns a copy of notes with the given locale's text set,
// adding the locale if it wasn't already present. It does not mutate notes.
func UpsertReleaseNotes(notes []*androidpublisher.LocalizedText, locale, text string) []*androidpublisher.LocalizedText {
	result := make([]*androidpublisher.LocalizedText, 0, len(notes)+1)
	found := false
	for _, n := range notes {
		if n.Language == locale {
			result = append(result, &androidpublisher.LocalizedText{Language: locale, Text: text})
			found = true
			continue
		}
		result = append(result, n)
	}
	if !found {
		result = append(result, &androidpublisher.LocalizedText{Language: locale, Text: text})
	}
	return result
}
