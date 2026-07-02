package playapi

import (
	"context"
	"fmt"
	"io"

	androidpublisher "google.golang.org/api/androidpublisher/v3"
)

// Image type values accepted by the Google Play Developer API. These are not
// exposed as Go constants by the SDK itself (the API treats imageType as a
// plain string) — they come from Google's REST API reference
// (developers.google.com/android-publisher/api-ref/rest/v3/edits.images).
const (
	ImageTypePhoneScreenshots     = "phoneScreenshots"
	ImageTypeSevenInchScreenshots = "sevenInchScreenshots"
	ImageTypeTenInchScreenshots   = "tenInchScreenshots"
	ImageTypeTVScreenshots        = "tvScreenshots"
	ImageTypeWearScreenshots      = "wearScreenshots"
	ImageTypeIcon                 = "icon"
	ImageTypeFeatureGraphic       = "featureGraphic"
	ImageTypePromoGraphic         = "promoGraphic"
	ImageTypeTVBanner             = "tvBanner"
)

// AllImageTypes lists every image type gpc knows how to pull/push.
var AllImageTypes = []string{
	ImageTypePhoneScreenshots,
	ImageTypeSevenInchScreenshots,
	ImageTypeTenInchScreenshots,
	ImageTypeTVScreenshots,
	ImageTypeWearScreenshots,
	ImageTypeIcon,
	ImageTypeFeatureGraphic,
	ImageTypePromoGraphic,
	ImageTypeTVBanner,
}

// ListImages returns the images of one locale + image type.
func (c *Client) ListImages(ctx context.Context, packageName, editID, language, imageType string) ([]*androidpublisher.Image, error) {
	resp, err := c.svc.Edits.Images.List(packageName, editID, language, imageType).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list %s/%s images for %s edit %s: %w", language, imageType, packageName, editID, err)
	}
	return resp.Images, nil
}

// UploadImage uploads one image (PNG/JPEG bytes from r) to a locale + type.
func (c *Client) UploadImage(ctx context.Context, packageName, editID, language, imageType string, r io.Reader) (*androidpublisher.Image, error) {
	resp, err := c.svc.Edits.Images.Upload(packageName, editID, language, imageType).Media(r).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("upload %s/%s image for %s edit %s: %w", language, imageType, packageName, editID, err)
	}
	return resp.Image, nil
}

// DeleteImage removes a single image by id.
func (c *Client) DeleteImage(ctx context.Context, packageName, editID, language, imageType, imageID string) error {
	if err := c.svc.Edits.Images.Delete(packageName, editID, language, imageType, imageID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete %s/%s image %s for %s edit %s: %w", language, imageType, imageID, packageName, editID, err)
	}
	return nil
}

// DeleteAllImages removes every image of a locale + type and returns them.
func (c *Client) DeleteAllImages(ctx context.Context, packageName, editID, language, imageType string) ([]*androidpublisher.Image, error) {
	resp, err := c.svc.Edits.Images.Deleteall(packageName, editID, language, imageType).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("delete all %s/%s images for %s edit %s: %w", language, imageType, packageName, editID, err)
	}
	return resp.Deleted, nil
}
