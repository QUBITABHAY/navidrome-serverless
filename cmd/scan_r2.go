package cmd

import (
	"os"

	"github.com/navidrome/navidrome/core/storage/r2"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/persistence/postgres"
	"github.com/spf13/cobra"
)

var scanR2Cmd = &cobra.Command{
	Use:   "scan-r2",
	Short: "Scan Cloudflare R2 bucket and synchronize metadata to Neon/PostgreSQL",
	Run: func(cmd *cobra.Command, args []string) {
		preRun()
		ctx := cmd.Context()

		pgURL := os.Getenv("NEON_DATABASE_URL")
		if pgURL == "" {
			pgURL = os.Getenv("POSTGRES_URL")
		}
		if pgURL == "" {
			pgURL = os.Getenv("ND_DATABASE_URL")
		}
		if pgURL == "" {
			log.Fatal(ctx, "NEON_DATABASE_URL or POSTGRES_URL environment variable must be set")
		}

		store, err := postgres.New(ctx, pgURL)
		if err != nil {
			log.Fatal(ctx, "Failed to connect to PostgreSQL", err)
		}
		defer store.Close()

		cfg := r2.ConfigFromGlobal()
		if cfg.Bucket == "" {
			log.Fatal(ctx, "ND_R2_BUCKET environment variable must be set")
		}

		log.Info(ctx, "Starting Cloudflare R2 library scan...", "bucket", cfg.Bucket)
		res, err := r2.SyncR2(ctx, store, cfg)
		if err != nil {
			log.Fatal(ctx, "R2 library scan failed", err)
		}

		log.Info(ctx, "R2 library scan completed successfully!",
			"totalObjects", res.TotalObjects,
			"audioFiles", res.AudioFiles,
			"processed", res.Processed,
			"errors", res.Errors,
			"duration", res.Duration.String(),
		)
	},
}

func init() {
	rootCmd.AddCommand(scanR2Cmd)
}
