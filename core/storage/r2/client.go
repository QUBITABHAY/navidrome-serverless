package r2

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/navidrome/navidrome/conf"
)

var (
	defaultClient        *s3.Client
	defaultPresignClient *s3.PresignClient
	clientOnce           sync.Once
	clientErr            error
)

// Config holds connection parameters for Cloudflare R2 / AWS S3.
type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Endpoint        string
	Region          string
	PublicURL       string
}

// ConfigFromGlobal returns the R2 configuration from Navidrome's server configuration.
func ConfigFromGlobal() Config {
	cfg := Config{
		AccountID:       conf.Server.R2.AccountID,
		AccessKeyID:     conf.Server.R2.AccessKeyID,
		SecretAccessKey: conf.Server.R2.SecretAccessKey,
		Bucket:          conf.Server.R2.Bucket,
		Endpoint:        conf.Server.R2.Endpoint,
		Region:          conf.Server.R2.Region,
		PublicURL:       conf.Server.R2.PublicURL,
	}
	if cfg.Bucket == "" {
		cfg.Bucket = os.Getenv("ND_R2_BUCKET")
	}
	if cfg.AccountID == "" {
		cfg.AccountID = os.Getenv("ND_R2_ACCOUNTID")
	}
	if cfg.AccessKeyID == "" {
		cfg.AccessKeyID = os.Getenv("ND_R2_ACCESSKEYID")
	}
	if cfg.SecretAccessKey == "" {
		cfg.SecretAccessKey = os.Getenv("ND_R2_SECRETACCESSKEY")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = os.Getenv("ND_R2_ENDPOINT")
	}
	if cfg.PublicURL == "" {
		cfg.PublicURL = os.Getenv("ND_R2_PUBLICURL")
	}
	if cfg.Region == "" {
		cfg.Region = "auto"
	}
	if cfg.Endpoint == "" && cfg.AccountID != "" {
		cfg.Endpoint = fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)
	}
	return cfg
}

// NewS3Client creates an AWS S3 client configured for Cloudflare R2 or standard S3.
func NewS3Client(ctx context.Context, cfg Config) (*s3.Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		if cfg.Endpoint != "" {
			return aws.Endpoint{
				URL:               cfg.Endpoint,
				SigningRegion:     cfg.Region,
				HostnameImmutable: true,
			}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})

	var optFns []func(*awsconfig.LoadOptions) error
	optFns = append(optFns, awsconfig.WithRegion(cfg.Region))

	if cfg.Endpoint != "" {
		optFns = append(optFns, awsconfig.WithEndpointResolverWithOptions(customResolver))
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		optFns = append(optFns, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		return nil, fmt.Errorf("loading aws config for r2: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true // Required for R2 and custom S3 endpoints
	})

	return client, nil
}

// DefaultClient returns the globally initialized S3/R2 client.
func DefaultClient(ctx context.Context) (*s3.Client, error) {
	clientOnce.Do(func() {
		cfg := ConfigFromGlobal()
		defaultClient, clientErr = NewS3Client(ctx, cfg)
		if clientErr == nil {
			defaultPresignClient = s3.NewPresignClient(defaultClient)
		}
	})
	return defaultClient, clientErr
}

// DefaultPresignClient returns the presign client.
func DefaultPresignClient(ctx context.Context) (*s3.PresignClient, error) {
	_, err := DefaultClient(ctx)
	if err != nil {
		return nil, err
	}
	return defaultPresignClient, nil
}

// CleanKey removes schema prefixes and leading slashes from a path.
func CleanKey(path string) string {
	path = strings.TrimPrefix(path, "r2://")
	path = strings.TrimPrefix(path, "s3://")
	path = strings.TrimPrefix(path, "/")
	return path
}
