package cmd

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/consts"
	"github.com/navidrome/navidrome/core/metrics"
	"github.com/navidrome/navidrome/log"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/persistence/postgres"
	"github.com/navidrome/navidrome/server"
	"github.com/navidrome/navidrome/server/events"
	"github.com/navidrome/navidrome/server/serverless"
)

var (
	serverlessApp  http.Handler
	serverlessOnce sync.Once
	serverlessErr  error
)

// GetServerlessApp initializes and returns the singleton serverless HTTP handler for Vercel.
func GetServerlessApp(ctx context.Context) (http.Handler, error) {
	serverlessOnce.Do(func() {
		// In serverless environments (e.g. Vercel, AWS Lambda), the filesystem is read-only except /tmp
		if os.Getenv("ND_DATAFOLDER") == "" {
			_ = os.Setenv("ND_DATAFOLDER", "/tmp")
		}
		if os.Getenv("ND_CACHEFOLDER") == "" {
			_ = os.Setenv("ND_CACHEFOLDER", "/tmp/cache")
		}
		if os.Getenv("ND_PLUGINS_ENABLED") == "" {
			_ = os.Setenv("ND_PLUGINS_ENABLED", "false")
		}

		conf.Load(true)

		conf.Server.DataFolder = conf.NewDir("/tmp")
		conf.Server.CacheFolder = conf.NewDir("/tmp/cache")
		conf.Server.Plugins.Enabled = false

		var ds model.DataStore
		var pgStore *postgres.PostgresStore
		pgURL := os.Getenv("DATABASE_URL")

		if pgURL != "" {
			log.Info(ctx, "Initializing Serverless DataStore with PostgreSQL")
			var err error
			pgStore, err = postgres.New(ctx, pgURL)
			if err != nil {
				serverlessErr = err
				return
			}
			ds = pgStore
		} else {
			log.Info(ctx, "Initializing Serverless DataStore with default persistence")
			ds = CreateDataStore()
		}

		broker := events.GetBroker()
		insights := metrics.GetInstance(ds)
		srv := server.New(ds, broker, insights)

		srv.MountRouter("Native API", consts.URLPathNativeAPI, CreateNativeAPIRouterWithDS(ctx, ds))
		srv.MountRouter("Subsonic API", consts.URLPathSubsonicAPI, CreateSubsonicAPIRouterWithDS(ctx, ds))
		srv.MountRouter("Public Endpoints", consts.URLPathPublic, CreatePublicRouterWithDS(ds))
		if pgStore != nil {
			srv.MountRouter("R2 Sync", "/api/scan/r2", serverless.ScanR2Handler(pgStore))
		}
		srv.MountWebUI()

		serverlessApp = srv
	})

	return serverlessApp, serverlessErr
}
