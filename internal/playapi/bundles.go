package playapi

import (
	"context"
	"fmt"
	"io"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// ListBundles returns the AABs already attached to the edit.
func (c *Client) ListBundles(ctx context.Context, packageName, editID string) ([]*androidpublisher.Bundle, error) {
	resp, err := c.svc.Edits.Bundles.List(packageName, editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list bundles for %s edit %s: %w", packageName, editID, err)
	}
	return resp.Bundles, nil
}

// UploadBundle uploads an .aab (bytes from r) into the edit. Google computes
// and returns the version code and hashes server-side.
func (c *Client) UploadBundle(ctx context.Context, packageName, editID string, r io.Reader) (*androidpublisher.Bundle, error) {
	bundle, err := c.svc.Edits.Bundles.Upload(packageName, editID).Media(r).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("upload bundle for %s edit %s: %w", packageName, editID, err)
	}
	return bundle, nil
}
