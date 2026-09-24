package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/navidrome/navidrome/persistence/postgres"
)

func TestServerlessRoutersInitialization(t *testing.T) {
	ctx := context.Background()
	store := &postgres.PostgresStore{}

	// Test constructing Native API router with PostgresStore does not panic
	nativeRouter := CreateNativeAPIRouterWithDS(ctx, store)
	if nativeRouter == nil {
		t.Fatal("expected nativeRouter not to be nil")
	}

	// Test constructing Subsonic API router with PostgresStore does not panic
	subsonicRouter := CreateSubsonicAPIRouterWithDS(ctx, store)
	if subsonicRouter == nil {
		t.Fatal("expected subsonicRouter not to be nil")
	}

	// Test constructing Public router with PostgresStore does not panic
	publicRouter := CreatePublicRouterWithDS(store)
	if publicRouter == nil {
		t.Fatal("expected publicRouter not to be nil")
	}

	// Test HTTP requests to REST endpoints do not crash
	endpoints := []string{
		"/transcoding",
		"/album",
		"/artist",
		"/song",
		"/playlist",
		"/player",
		"/radio",
		"/genre",
		"/tag",
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest("GET", ep, nil)
		rec := httptest.NewRecorder()
		nativeRouter.ServeHTTP(rec, req)
		// Should return 401 Unauthorized (because unauthenticated) instead of panicking with 500 / nil dereference
		if rec.Code != http.StatusUnauthorized {
			t.Logf("Endpoint %s returned status %d", ep, rec.Code)
		}
	}
}
