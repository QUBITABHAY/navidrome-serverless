package handler

import (
	"net/http"
	"strings"

	"github.com/navidrome/navidrome/cmd"
)

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	// If Vercel rewrote the request to /api, recover the original client path
	if r.URL.Path == "/api" || r.URL.Path == "/api/" || r.URL.Path == "/api/index" || r.URL.Path == "/api/index.go" {
		if matched := r.Header.Get("X-Matched-Path"); matched != "" {
			r.URL.Path = matched
		} else if fwd := r.Header.Get("X-Forwarded-Uri"); fwd != "" {
			if idx := strings.Index(fwd, "?"); idx != -1 {
				r.URL.Path = fwd[:idx]
			} else {
				r.URL.Path = fwd
			}
		}
	}

	app, err := cmd.GetServerlessApp(r.Context())
	if err != nil {
		http.Error(w, "Failed to initialize Navidrome serverless application: "+err.Error(), http.StatusInternalServerError)
		return
	}
	app.ServeHTTP(w, r)
}
