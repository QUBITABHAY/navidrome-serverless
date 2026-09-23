package r2

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/navidrome/navidrome/core/storage"
	"github.com/navidrome/navidrome/log"
)

type R2Storage struct {
	u      url.URL
	client *s3.Client
	bucket string
	prefix string
}

func init() {
	storage.Register("r2", newR2Storage)
	storage.Register("s3", newR2Storage)
}

func newR2Storage(u url.URL) storage.Storage {
	bucket := u.Host
	prefix := strings.TrimPrefix(u.Path, "/")

	ctx := context.Background()
	client, err := DefaultClient(ctx)
	if err != nil {
		log.Error("Failed to initialize R2 storage client", "url", u.String(), err)
	}

	return &R2Storage{
		u:      u,
		client: client,
		bucket: bucket,
		prefix: prefix,
	}
}

func (s *R2Storage) FS() (storage.MusicFS, error) {
	if s.client == nil {
		ctx := context.Background()
		client, err := DefaultClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("initializing r2 client: %w", err)
		}
		s.client = client
	}

	return newR2FS(context.Background(), s.client, s.bucket, s.prefix), nil
}

var _ storage.Storage = (*R2Storage)(nil)
