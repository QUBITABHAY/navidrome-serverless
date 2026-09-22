package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/navidrome/navidrome/model"
	"github.com/navidrome/navidrome/persistence/postgres/pgdb"
)

func TestToModelUser(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	dbUser := pgdb.User{
		ID:             "u-123",
		UserName:       "testuser",
		Name:           "Test User",
		Email:          "test@example.com",
		Password:       "hashed_password",
		IsAdmin:        true,
		TokenEpoch:     1,
		ScrobbleFilter: "",
		LastLoginAt:    pgtype.Timestamptz{Time: now, Valid: true},
		LastAccessAt:   pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:      pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:      pgtype.Timestamptz{Time: now, Valid: true},
	}

	u := toModelUser(dbUser)
	if u.ID != "u-123" {
		t.Errorf("expected ID u-123, got %s", u.ID)
	}
	if u.UserName != "testuser" {
		t.Errorf("expected UserName testuser, got %s", u.UserName)
	}
	if !u.IsAdmin {
		t.Errorf("expected IsAdmin true, got false")
	}
	if u.LastLoginAt == nil || !u.LastLoginAt.Equal(now) {
		t.Errorf("expected LastLoginAt %v, got %v", now, u.LastLoginAt)
	}
}

func TestToModelLibrary(t *testing.T) {
	now := time.Now().Truncate(time.Microsecond)
	dbLib := pgdb.Library{
		ID:                 1,
		Name:               "Music Library",
		Path:               "r2://bucket/music",
		RemotePath:         "",
		LastScanAt:         pgtype.Timestamptz{Time: now, Valid: true},
		LastScanStartedAt:  pgtype.Timestamptz{Time: now, Valid: true},
		FullScanInProgress: false,
		TotalSongs:         42,
		TotalAlbums:        3,
		TotalArtists:       2,
		TotalFolders:       4,
		TotalFiles:         42,
		TotalMissingFiles:  0,
		TotalSize:          1048576,
		TotalDuration:      180.5,
		DefaultNewUsers:    true,
		CreatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:          pgtype.Timestamptz{Time: now, Valid: true},
	}

	l := toModelLibrary(dbLib)
	if l.ID != 1 {
		t.Errorf("expected ID 1, got %d", l.ID)
	}
	if l.Name != "Music Library" {
		t.Errorf("expected Name 'Music Library', got %s", l.Name)
	}
	if l.TotalSongs != 42 {
		t.Errorf("expected TotalSongs 42, got %d", l.TotalSongs)
	}
	if !l.DefaultNewUsers {
		t.Errorf("expected DefaultNewUsers true, got false")
	}
}

func TestPostgresStore_LiveConnection(t *testing.T) {
	connStr := os.Getenv("TEST_POSTGRES_URL")
	if connStr == "" {
		t.Skip("Skipping live PostgreSQL test (set TEST_POSTGRES_URL to run)")
	}

	ctx := context.Background()
	store, err := New(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}
	defer store.Close()

	// Test Property repository
	props := store.Property(ctx)
	if err := props.Put("test_key", "test_value"); err != nil {
		t.Fatalf("failed to put property: %v", err)
	}

	val, err := props.Get("test_key")
	if err != nil {
		t.Fatalf("failed to get property: %v", err)
	}
	if val != "test_value" {
		t.Errorf("expected 'test_value', got %s", val)
	}

	// Test Library repository
	libs := store.Library(ctx)
	lib := &model.Library{
		ID:   1,
		Name: "Test Music",
		Path: "r2://my-bucket/music",
	}
	if err := libs.Put(lib); err != nil {
		t.Fatalf("failed to put library: %v", err)
	}

	loadedLib, err := libs.Get(1)
	if err != nil {
		t.Fatalf("failed to get library: %v", err)
	}
	if loadedLib.Name != "Test Music" {
		t.Errorf("expected library name 'Test Music', got %s", loadedLib.Name)
	}
}
