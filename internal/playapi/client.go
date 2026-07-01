// Package playapi is gpc's service layer over the official androidpublisher
// SDK. Command packages depend only on this package, never on the SDK types'
// call builders, which keeps SDK churn contained and testing uniform.
package playapi

import (
	"context"
	"fmt"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// Client wraps an androidpublisher.Service and is the only type command
// packages should depend on for talking to the Google Play Developer API.
type Client struct {
	svc *androidpublisher.Service
}

// NewClient wraps an already-authenticated androidpublisher.Service.
func NewClient(svc *androidpublisher.Service) *Client {
	return &Client{svc: svc}
}

// BeginEdit opens a new edit transaction for the app.
func (c *Client) BeginEdit(ctx context.Context, packageName string) (*androidpublisher.AppEdit, error) {
	edit, err := c.svc.Edits.Insert(packageName, &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("begin edit for %s: %w", packageName, err)
	}
	return edit, nil
}

// CommitEdit commits the edit, making its changes live in Play Console.
func (c *Client) CommitEdit(ctx context.Context, packageName, editID string) (*androidpublisher.AppEdit, error) {
	edit, err := c.svc.Edits.Commit(packageName, editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("commit edit %s for %s: %w", editID, packageName, err)
	}
	return edit, nil
}

// DiscardEdit deletes the edit without applying its changes.
func (c *Client) DiscardEdit(ctx context.Context, packageName, editID string) error {
	if err := c.svc.Edits.Delete(packageName, editID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("discard edit %s for %s: %w", editID, packageName, err)
	}
	return nil
}

// GetEdit fetches an existing edit (useful to check expiry).
func (c *Client) GetEdit(ctx context.Context, packageName, editID string) (*androidpublisher.AppEdit, error) {
	edit, err := c.svc.Edits.Get(packageName, editID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get edit %s for %s: %w", editID, packageName, err)
	}
	return edit, nil
}

// WithTransaction runs fn against an edit transaction. If editID is already
// set (the caller passed --edit-id to share a transaction across commands),
// fn runs against that edit and the caller remains responsible for
// committing it later via `gpc edits commit`. Otherwise a new edit is begun,
// committed on success, and discarded if fn returns an error.
func (c *Client) WithTransaction(ctx context.Context, packageName, editID string, fn func(editID string) error) error {
	if editID != "" {
		return fn(editID)
	}
	edit, err := c.BeginEdit(ctx, packageName)
	if err != nil {
		return err
	}
	if err := fn(edit.Id); err != nil {
		_ = c.DiscardEdit(ctx, packageName, edit.Id)
		return err
	}
	if _, err := c.CommitEdit(ctx, packageName, edit.Id); err != nil {
		return err
	}
	return nil
}

// WithReadOnlyEdit begins an edit, runs fn, then always discards the edit.
// Use it for read paths that the API only exposes inside an edit (AppDetails,
// listings) so no accidental commit can occur.
func (c *Client) WithReadOnlyEdit(ctx context.Context, packageName string, fn func(editID string) error) error {
	edit, err := c.BeginEdit(ctx, packageName)
	if err != nil {
		return err
	}
	defer func() { _ = c.DiscardEdit(ctx, packageName, edit.Id) }()
	return fn(edit.Id)
}
