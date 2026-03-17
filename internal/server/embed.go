package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:ui
var uiFS embed.FS

// embeddedUI returns an http.Handler serving the embedded Svelte build.
func embeddedUI() http.Handler {
	sub, err := fs.Sub(uiFS, "ui")
	if err != nil {
		panic("failed to get embedded UI subtree: " + err.Error())
	}
	fileServer := http.FileServer(http.FS(sub))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists in embedded FS
		f, err := sub.Open(path[1:]) // strip leading /
		if err != nil {
			// SPA fallback: serve index.html for client-side routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}
