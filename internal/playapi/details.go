package playapi

import (
	"context"
	"fmt"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// GetDetails fetches app-level details (default language, contact info).
func (c *Client) GetDetails(ctx context.Context, packageName, editID string) (*androidpublisher.AppDetails, error) {
	details, err := c.svc.Edits.Details.Get(packageName, editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get details for %s edit %s: %w", packageName, editID, err)
	}
	return details, nil
}

// UpdateDetails replaces app-level details.
func (c *Client) UpdateDetails(ctx context.Context, packageName, editID string, details *androidpublisher.AppDetails) (*androidpublisher.AppDetails, error) {
	updated, err := c.svc.Edits.Details.Update(packageName, editID, details).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("update details for %s edit %s: %w", packageName, editID, err)
	}
	return updated, nil
}
