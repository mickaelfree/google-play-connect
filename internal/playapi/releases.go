package playapi

import (
	"context"
	"fmt"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// ListReleaseSummaries returns the active releases of a track WITHOUT opening
// an edit — this is the only read-only release endpoint in the API and powers
// `gpc status`. The API caps the response at 20 releases.
func (c *Client) ListReleaseSummaries(ctx context.Context, packageName, track string) ([]*androidpublisher.ReleaseSummary, error) {
	parent := fmt.Sprintf("applications/%s/tracks/%s", packageName, track)
	resp, err := c.svc.Applications.Tracks.Releases.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list release summaries for %s track %s: %w", packageName, track, err)
	}
	return resp.Releases, nil
}
