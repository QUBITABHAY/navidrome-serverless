package handler

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/navidrome/navidrome/cmd"
)

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			buf := make([]byte, 8192)
			n := runtime.Stack(buf, false)
			stack := string(buf[:n])
			http.Error(w, fmt.Sprintf("SERVERLESS PANIC: %v\n\nStack:\n%s", rec, stack), http.StatusInternalServerError)
		}
	}()
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
