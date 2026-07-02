package playapi

import (
	"context"
	"fmt"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// ListListings returns all localized store listings in the edit.
func (c *Client) ListListings(ctx context.Context, packageName, editID string) ([]*androidpublisher.Listing, error) {
	resp, err := c.svc.Edits.Listings.List(packageName, editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list listings for %s edit %s: %w", packageName, editID, err)
	}
	return resp.Listings, nil
}

// GetListing returns one locale's store listing.
func (c *Client) GetListing(ctx context.Context, packageName, editID, language string) (*androidpublisher.Listing, error) {
	listing, err := c.svc.Edits.Listings.Get(packageName, editID, language).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get %s listing for %s edit %s: %w", language, packageName, editID, err)
	}
	return listing, nil
}

// UpdateListing creates or replaces one locale's store listing.
func (c *Client) UpdateListing(ctx context.Context, packageName, editID, language string, listing *androidpublisher.Listing) (*androidpublisher.Listing, error) {
	updated, err := c.svc.Edits.Listings.Update(packageName, editID, language, listing).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("update %s listing for %s edit %s: %w", language, packageName, editID, err)
	}
	return updated, nil
}

// DeleteListing removes one locale's store listing.
func (c *Client) DeleteListing(ctx context.Context, packageName, editID, language string) error {
	if err := c.svc.Edits.Listings.Delete(packageName, editID, language).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete %s listing for %s edit %s: %w", language, packageName, editID, err)
	}
	return nil
}
