package serverless

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/navidrome/navidrome/core/storage/r2"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/persistence/postgres"
)

// ScanR2Handler exposes an HTTP endpoint to trigger on-demand synchronization of the Cloudflare R2 bucket.
func ScanR2Handler(store *postgres.PostgresStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Security: verify webhook secret if configured
		secret := os.Getenv("ND_SCAN_SECRET")
		if secret != "" {
			authHeader := r.Header.Get("Authorization")
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				token = r.Header.Get("X-Scan-Secret")
			}
			if token == "" {
				token = r.URL.Query().Get("secret")
			}

			if token != secret {
				http.Error(w, "Unauthorized: invalid scan secret", http.StatusUnauthorized)
				return
			}
		}

		ctx := r.Context()
		cfg := r2.ConfigFromGlobal()
		if cfg.Bucket == "" {
			http.Error(w, "Cloudflare R2 bucket is not configured (set ND_R2_BUCKET)", http.StatusBadRequest)
			return
		}

		log.Info(ctx, "Starting on-demand Cloudflare R2 sync", "bucket", cfg.Bucket)

		result, err := r2.SyncR2(ctx, store, cfg)
		if err != nil {
			log.Error(ctx, "Cloudflare R2 sync failed", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":       "success",
			"totalObjects": result.TotalObjects,
			"audioFiles":   result.AudioFiles,
			"processed":    result.Processed,
			"errors":       result.Errors,
			"duration":     result.Duration.String(),
		})
	}
}
