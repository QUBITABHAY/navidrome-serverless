package r2

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type r2FileInfo struct {
	name    string
	size    int64
	modTime time.Time
	isDir   bool
}

func (fi r2FileInfo) Name() string         { return fi.name }
func (fi r2FileInfo) Size() int64          { return fi.size }
func (fi r2FileInfo) Mode() fs.FileMode    { return 0o644 }
func (fi r2FileInfo) ModTime() time.Time   { return fi.modTime }
func (fi r2FileInfo) IsDir() bool          { return fi.isDir }
func (fi r2FileInfo) Sys() interface{}     { return nil }
func (fi r2FileInfo) BirthTime() time.Time { return fi.modTime }

// r2File implements fs.File, io.Seeker, io.ReaderAt, and io.Closer over S3/R2 object storage.
type r2File struct {
	ctx      context.Context
	client   *s3.Client
	bucket   string
	key      string
	info     r2FileInfo
	offset   int64
	body     io.ReadCloser
	bodyLock sync.Mutex
}

func newR2File(ctx context.Context, client *s3.Client, bucket, key string) (*r2File, error) {
	head, err := client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("head object %s/%s: %w", bucket, key, err)
	}

	var modTime time.Time
	if head.LastModified != nil {
		modTime = *head.LastModified
	}
	var size int64
	if head.ContentLength != nil {
		size = *head.ContentLength
	}

	return &r2File{
		ctx:    ctx,
		client: client,
		bucket: bucket,
		key:    key,
		info: r2FileInfo{
			name:    path.Base(key),
			size:    size,
			modTime: modTime,
			isDir:   false,
		},
		offset: 0,
	}, nil
}

func (f *r2File) Stat() (fs.FileInfo, error) {
	return f.info, nil
}

func (f *r2File) Read(p []byte) (int, error) {
	if f.offset >= f.info.size {
		return 0, io.EOF
	}

	f.bodyLock.Lock()
	defer f.bodyLock.Unlock()

	// Read up to len(p) using S3 range request
	end := f.offset + int64(len(p)) - 1
	if end >= f.info.size {
		end = f.info.size - 1
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", f.offset, end)
	resp, err := f.client.GetObject(f.ctx, &s3.GetObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(f.key),
		Range:  aws.String(rangeHeader),
	})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	n, err := io.ReadFull(resp.Body, p[:end-f.offset+1])
	f.offset += int64(n)
	return n, err
}

func (f *r2File) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = f.offset + offset
	case io.SeekEnd:
		newOffset = f.info.size + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative seek position: %d", newOffset)
	}
	f.offset = newOffset
	return f.offset, nil
}

func (f *r2File) ReadAt(p []byte, off int64) (int, error) {
	if off >= f.info.size {
		return 0, io.EOF
	}
	end := off + int64(len(p)) - 1
	if end >= f.info.size {
		end = f.info.size - 1
	}

	rangeHeader := fmt.Sprintf("bytes=%d-%d", off, end)
	resp, err := f.client.GetObject(f.ctx, &s3.GetObjectInput{
		Bucket: aws.String(f.bucket),
		Key:    aws.String(f.key),
		Range:  aws.String(rangeHeader),
	})
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	return io.ReadFull(resp.Body, p[:end-off+1])
}

func (f *r2File) Close() error {
	f.bodyLock.Lock()
	defer f.bodyLock.Unlock()
	if f.body != nil {
		err := f.body.Close()
		f.body = nil
		return err
	}
	return nil
}

var _ fs.File = (*r2File)(nil)
var _ io.Seeker = (*r2File)(nil)
var _ io.ReaderAt = (*r2File)(nil)
