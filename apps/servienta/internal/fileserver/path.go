package fileserver

import "path/filepath"

// securePath resolves name under root without letting it escape (R6: names
// are opaque, but the tree boundary is absolute).
func securePath(root, name string) string {
	return filepath.Join(root, filepath.Clean("/"+name))
}
