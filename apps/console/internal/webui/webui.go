// Package webui embeds the built Vue SPA. The dist/ directory is produced by
// the app build and embedded at compile time so the console ships as one
// self-contained binary.
package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// FS returns the SPA files rooted at dist/.
func FS() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
