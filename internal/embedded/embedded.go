package embedded

import "embed"

// FS holds all hook and skill files baked into the ptk binary at build time.
// ptk init extracts these — no external hooks/ directory required after install.
//
//go:embed hooks skills
var FS embed.FS
