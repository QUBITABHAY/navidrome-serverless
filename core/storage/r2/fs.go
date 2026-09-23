package r2

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/navidrome/navidrome/core/storage"
	"github.com/navidrome/navidrome/model/metadata"
)

type r2FS struct {
	ctx    context.Context
	client *s3.Client
	bucket string
	prefix string
}

func newR2FS(ctx context.Context, client *s3.Client, bucket, prefix string) *r2FS {
	return &r2FS{
		ctx:    ctx,
		client: client,
		bucket: bucket,
		prefix: strings.TrimPrefix(prefix, "/"),
	}
}

func (rfs *r2FS) fullKey(name string) string {
	name = strings.TrimPrefix(name, "/")
	if rfs.prefix == "" {
		return name
	}
	return path.Join(rfs.prefix, name)
}

func (rfs *r2FS) Open(name string) (fs.File, error) {
	key := rfs.fullKey(name)
	return newR2File(rfs.ctx, rfs.client, rfs.bucket, key)
}

func (rfs *r2FS) ReadTags(paths ...string) (map[string]metadata.Info, error) {
	result := make(map[string]metadata.Info, len(paths))

	for _, p := range paths {
		key := rfs.fullKey(p)
		head, err := rfs.client.HeadObject(rfs.ctx, &s3.HeadObjectInput{
			Bucket: aws.String(rfs.bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return nil, fmt.Errorf("reading tags for %s: %w", key, err)
		}

		var modTime time.Time
		if head.LastModified != nil {
			modTime = *head.LastModified
		}
		var size int64
		if head.ContentLength != nil {
			size = *head.ContentLength
		}

		fi := r2FileInfo{
			name:    path.Base(key),
			size:    size,
			modTime: modTime,
			isDir:   false,
		}

		// Extract basic file metadata from R2 object properties
		info := metadata.Info{
			FileInfo: fi,
			AudioProperties: metadata.AudioProperties{
				Duration:   0,
				BitRate:    128,
				SampleRate: 44100,
				Channels:   2,
				Codec:      strings.TrimPrefix(path.Ext(key), "."),
			},
		}

		result[p] = info
	}

	return result, nil
}

var _ storage.MusicFS = (*r2FS)(nil)
