// Package skillsfs embeds the agent skills shipped with gpc.
package skillsfs

import "embed"

// FS holds skills/<name>/SKILL.md files.
//
//go:embed skills
var FS embed.FS
