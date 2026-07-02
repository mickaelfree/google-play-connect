package metadata

import (
	"fmt"
	"image"
	_ "image/jpeg" // register decoder for image.DecodeConfig
	_ "image/png"  // register decoder for image.DecodeConfig
	"os"
	"unicode/utf8"
)

// Google Play listing limits (characters, per Play Console help).
const (
	MaxTitleLen            = 30
	MaxShortDescriptionLen = 80
	MaxFullDescriptionLen  = 4000
)

// Google Play image dimension rules (pixels, per Play Console help).
const (
	IconSize             = 512
	FeatureGraphicWidth  = 1024
	FeatureGraphicHeight = 500
	PromoGraphicWidth    = 180
	PromoGraphicHeight   = 120
	TVBannerWidth        = 1280
	TVBannerHeight       = 720
	MinScreenshotSide    = 320
	MaxScreenshotSide    = 3840
)

// Issue is one offline-validation finding.
type Issue struct {
	Locale  string `json:"locale"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateListing checks one locale's listing against Play limits.
func ValidateListing(locale string, l Listing) []Issue {
	var issues []Issue
	add := func(field, message string) {
		issues = append(issues, Issue{Locale: locale, Field: field, Message: message})
	}
	if l.Title == "" {
		add("title", "title is required")
	} else if n := utf8.RuneCountInString(l.Title); n > MaxTitleLen {
		add("title", fmt.Sprintf("title is %d chars, max %d", n, MaxTitleLen))
	}
	if n := utf8.RuneCountInString(l.ShortDescription); n > MaxShortDescriptionLen {
		add("short_description", fmt.Sprintf("short description is %d chars, max %d", n, MaxShortDescriptionLen))
	}
	if n := utf8.RuneCountInString(l.FullDescription); n > MaxFullDescriptionLen {
		add("full_description", fmt.Sprintf("full description is %d chars, max %d", n, MaxFullDescriptionLen))
	}
	return issues
}

// exactSizes describes the required dimensions of single-image types.
var exactSizes = map[string][2]int{
	"icon":           {IconSize, IconSize},
	"featureGraphic": {FeatureGraphicWidth, FeatureGraphicHeight},
	"promoGraphic":   {PromoGraphicWidth, PromoGraphicHeight},
	"tvBanner":       {TVBannerWidth, TVBannerHeight},
}

// validateImageFile decodes only the image header and checks format and
// dimensions against the rules for its image type.
func validateImageFile(locale, imageType, path string) []Issue {
	f, err := os.Open(path)
	if err != nil {
		return []Issue{{Locale: locale, Field: imageType, Message: fmt.Sprintf("%s: %v", path, err)}}
	}
	defer f.Close()
	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return []Issue{{Locale: locale, Field: imageType, Message: fmt.Sprintf("%s: not a decodable PNG/JPEG image", path)}}
	}
	if format != "png" && format != "jpeg" {
		return []Issue{{Locale: locale, Field: imageType, Message: fmt.Sprintf("%s: format %s not allowed (png/jpeg only)", path, format)}}
	}
	if want, ok := exactSizes[imageType]; ok {
		if cfg.Width != want[0] || cfg.Height != want[1] {
			return []Issue{{Locale: locale, Field: imageType, Message: fmt.Sprintf("%s: is %dx%d, must be exactly %dx%d", path, cfg.Width, cfg.Height, want[0], want[1])}}
		}
		return nil
	}
	// Screenshot types: each side within [MinScreenshotSide, MaxScreenshotSide].
	if cfg.Width < MinScreenshotSide || cfg.Height < MinScreenshotSide ||
		cfg.Width > MaxScreenshotSide || cfg.Height > MaxScreenshotSide {
		return []Issue{{Locale: locale, Field: imageType, Message: fmt.Sprintf("%s: is %dx%d, each side must be within %d-%d px", path, cfg.Width, cfg.Height, MinScreenshotSide, MaxScreenshotSide)}}
	}
	return nil
}

// ValidateImages checks every local image file against Play format and
// dimension rules, without any API call.
func ValidateImages(root, pkg string) ([]Issue, error) {
	locales, err := ImageLocales(root, pkg)
	if err != nil {
		return nil, err
	}
	var issues []Issue
	allTypes := append(append([]string{}, ScreenshotTypes...), SingleImageTypes...)
	for _, locale := range locales {
		for _, imageType := range allTypes {
			paths, err := LocalImages(root, pkg, locale, imageType)
			if err != nil {
				return nil, err
			}
			for _, path := range paths {
				issues = append(issues, validateImageFile(locale, imageType, path)...)
			}
		}
	}
	return issues, nil
}

// ValidateTree validates every locale listing and every local image under
// root for the app.
func ValidateTree(root, pkg string) ([]Issue, error) {
	locales, err := ListingLocales(root, pkg)
	if err != nil {
		return nil, err
	}
	var issues []Issue
	for _, locale := range locales {
		l, err := ReadListing(root, pkg, locale)
		if err != nil {
			return nil, err
		}
		issues = append(issues, ValidateListing(locale, l)...)
	}
	imageIssues, err := ValidateImages(root, pkg)
	if err != nil {
		return nil, err
	}
	return append(issues, imageIssues...), nil
}
