package r2

import (
	"context"
	"testing"
	"time"

	"github.com/navidrome/navidrome/conf"
	"github.com/navidrome/navidrome/core/storage"
)

func TestCleanKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"r2://bucket/music/track.mp3", "bucket/music/track.mp3"},
		{"s3://bucket/music/track.mp3", "bucket/music/track.mp3"},
		{"/music/track.mp3", "music/track.mp3"},
		{"music/track.mp3", "music/track.mp3"},
	}

	for _, tt := range tests {
		got := CleanKey(tt.input)
		if got != tt.expected {
			t.Errorf("CleanKey(%q) = %q, expected %q", tt.input, got, tt.expected)
		}
	}
}

func TestSplitBucketKey(t *testing.T) {
	conf.Server.R2.Bucket = "default-bucket"

	tests := []struct {
		input          string
		expectedBucket string
		expectedKey    string
	}{
		{"r2://my-bucket/artist/album/song.mp3", "my-bucket", "artist/album/song.mp3"},
		{"s3://cloud-music/album/song.flac", "cloud-music", "album/song.flac"},
		{"artist/album/song.mp3", "default-bucket", "artist/album/song.mp3"},
		{"/artist/album/song.mp3", "default-bucket", "artist/album/song.mp3"},
	}

	for _, tt := range tests {
		b, k := SplitBucketKey(tt.input)
		if b != tt.expectedBucket || k != tt.expectedKey {
			t.Errorf("SplitBucketKey(%q) = (%q, %q), expected (%q, %q)",
				tt.input, b, k, tt.expectedBucket, tt.expectedKey)
		}
	}
}

func TestIsR2Path(t *testing.T) {
	conf.Server.R2.EnablePresignedStream = false
	conf.Server.R2.Bucket = ""

	if !IsR2Path("r2://my-bucket/song.mp3") {
		t.Errorf("expected r2:// to be recognized as R2 path")
	}
	if !IsR2Path("s3://my-bucket/song.mp3") {
		t.Errorf("expected s3:// to be recognized as R2 path")
	}
	if IsR2Path("/local/path/song.mp3") {
		t.Errorf("expected local path NOT to be recognized when presigned stream is disabled")
	}

	conf.Server.R2.EnablePresignedStream = true
	conf.Server.R2.Bucket = "my-bucket"
	if !IsR2Path("music/song.mp3") {
		t.Errorf("expected relative path to be recognized when presigned stream is enabled with bucket")
	}
}

func TestPresignGet_PublicURL(t *testing.T) {
	conf.Server.R2.Bucket = "my-bucket"
	conf.Server.R2.PublicURL = "https://pub-abc.r2.dev"

	ctx := context.Background()
	url, err := PresignGet(ctx, "r2://my-bucket/music/song.mp3", 1*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "https://pub-abc.r2.dev/music/song.mp3"
	if url != expected {
		t.Errorf("expected %q, got %q", expected, url)
	}
}

func TestStorageRegistration(t *testing.T) {
	sR2, err := storage.For("r2://music-bucket/library")
	if err != nil {
		t.Fatalf("expected r2 schema to be registered, got error: %v", err)
	}
	if sR2 == nil {
		t.Fatalf("expected non-nil Storage for r2://")
	}

	sS3, err := storage.For("s3://music-bucket/library")
	if err != nil {
		t.Fatalf("expected s3 schema to be registered, got error: %v", err)
	}
	if sS3 == nil {
		t.Fatalf("expected non-nil Storage for s3://")
	}
}
