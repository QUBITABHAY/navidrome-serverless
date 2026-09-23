package handler

import (
	"net/http"

	"github.com/navidrome/navidrome/cmd"
)

// Handler is the Vercel serverless function entry point
func Handler(w http.ResponseWriter, r *http.Request) {
	app, err := cmd.GetServerlessApp(r.Context())
	if err != nil {
		http.Error(w, "Failed to initialize Navidrome serverless application: "+err.Error(), http.StatusInternalServerError)
		return
	}
	app.ServeHTTP(w, r)
}
