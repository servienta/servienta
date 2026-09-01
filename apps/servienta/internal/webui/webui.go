// Package webui embeds the built management SPA (webui/ at the module root).
// The dist/ directory is produced by the SPA build; the engine serves it on
// the control port so the delivered container carries its own UI and guide
// (D20). Built without the SPA (e.g. plain go test), dist holds only a
// placeholder and the UI is simply absent — the contract is unaffected.
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
