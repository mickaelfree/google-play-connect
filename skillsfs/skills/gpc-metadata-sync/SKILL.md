---
name: gpc-metadata-sync
description: Use when editing Google Play store listings (title, descriptions, screenshots) across locales with gpc — pull the current listing tree, edit files locally, validate offline, push back in one transaction.
---

# gpc Metadata Sync

## Workflow

```bash
# 1. Pull current listings + image manifests
gpc metadata pull --app com.example.app --dir ./metadata

# 2. Edit the files
#    metadata/com.example.app/listings/<locale>/title.txt              (max 30 chars)
#    metadata/com.example.app/listings/<locale>/short_description.txt (max 80 chars)
#    metadata/com.example.app/listings/<locale>/full_description.txt  (max 4000 chars)
#    metadata/com.example.app/listings/<locale>/video.txt             (YouTube URL)
#    metadata/com.example.app/images/<locale>/phoneScreenshots/1.png, 2.png...
#    metadata/com.example.app/images/<locale>/icon.png, featureGraphic.png

# 3. Validate offline (no API calls, catches limit violations)
gpc metadata validate --app com.example.app --dir ./metadata

# 4. Push everything in ONE edit transaction
gpc metadata push --app com.example.app --dir ./metadata --confirm
```

## Rules

- Adding a locale = create its `listings/<locale>/` directory with at least
  `title.txt`. Push creates the listing.
- Image push is destructive per locale/type: local files REPLACE all remote
  images of that type (delete-all + upload, sorted by file name). A
  locale/type with no local files is left untouched.
- `*.manifest.json` files are pull artifacts (remote ids/urls); they are
  ignored by push — do not edit them.
- To batch metadata with a release in one commit: `gpc edits begin`, pass
  `--edit-id` to push and to `gpc releases publish`, then `gpc edits commit`.
