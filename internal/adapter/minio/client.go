package minioadapter

import (
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// publicReadPolicy returns a minimal S3 bucket policy allowing anonymous GET.
func publicReadPolicy(bucket string) string {
	return fmt.Sprintf(`{
  "Version":"2012-10-17",
  "Statement":[
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetBucketLocation","s3:ListBucket"],
      "Resource":["arn:aws:s3:::%s"]
    },
    {
      "Effect":"Allow",
      "Principal":{"AWS":["*"]},
      "Action":["s3:GetObject"],
      "Resource":["arn:aws:s3:::%s/*"]
    }
  ]
}`, bucket, bucket)
}

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
		// Only the main bucket (not KYC) should be publicly readable.
		if !strings.HasSuffix(b, "-kyc") {
			_ = client.SetBucketPolicy(ctx, b, publicReadPolicy(b))
		}
	}

	return client, nil
}
