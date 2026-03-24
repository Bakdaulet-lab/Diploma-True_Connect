package minioadapter

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// New creates a new MinIO client and ensures the default bucket exists.
func New(ctx context.Context, endpoint, accessKey, secretKey string, useSSL bool, bucket string) (*minio.Client, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("creating minio client: %w", err)
	}

	// Ensure buckets exist.
	buckets := []string{bucket, bucket + "-kyc"}
	for _, b := range buckets {
		exists, err := client.BucketExists(ctx, b)
		if err != nil {
			return nil, fmt.Errorf("checking bucket %q: %w", b, err)
		}
		if !exists {
			if err := client.MakeBucket(ctx, b, minio.MakeBucketOptions{}); err != nil {
				return nil, fmt.Errorf("creating bucket %q: %w", b, err)
			}
		}
	}

	return client, nil
}
