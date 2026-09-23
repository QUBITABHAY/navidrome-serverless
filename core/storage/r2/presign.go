package r2

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/navidrome/navidrome/conf"
)

// IsR2Path reports whether a given path is an R2/S3 object path or if R2 streaming is active.
func IsR2Path(path string) bool {
	if strings.HasPrefix(path, "r2://") || strings.HasPrefix(path, "s3://") {
		return true
	}
	return conf.Server.R2.EnablePresignedStream && conf.Server.R2.Bucket != ""
}

// SplitBucketKey parses a path or URI into bucket and object key.
func SplitBucketKey(rawPath string) (bucket string, key string) {
	clean := CleanKey(rawPath)
	if strings.HasPrefix(rawPath, "r2://") || strings.HasPrefix(rawPath, "s3://") {
		// Format: r2://my-bucket/path/to/song.mp3
		parts := strings.SplitN(clean, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1]
		}
		return parts[0], ""
	}

	// If no bucket in path, use configured bucket from config
	bucket = conf.Server.R2.Bucket
	key = clean
	return bucket, key
}

// PresignGet generates a presigned GET URL for an audio file in Cloudflare R2 / AWS S3.
func PresignGet(ctx context.Context, keyOrPath string, expiry time.Duration) (string, error) {
	bucket, key := SplitBucketKey(keyOrPath)
	if bucket == "" || key == "" {
		return "", fmt.Errorf("invalid r2 path: bucket=%q, key=%q", bucket, key)
	}

	// 1. If a public CDN URL is configured (e.g., https://pub-xxx.r2.dev or https://music.mycdn.com)
	if conf.Server.R2.PublicURL != "" {
		baseURL := strings.TrimSuffix(conf.Server.R2.PublicURL, "/")
		return fmt.Sprintf("%s/%s", baseURL, key), nil
	}

	// 2. Otherwise generate a cryptographically signed presigned S3/R2 URL
	presigner, err := DefaultPresignClient(ctx)
	if err != nil {
		return "", fmt.Errorf("obtaining presigner: %w", err)
	}

	if expiry <= 0 {
		expiry = 1 * time.Hour
	}

	req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return "", fmt.Errorf("presigning s3 get object for %s/%s: %w", bucket, key, err)
	}

	return req.URL, nil
}
