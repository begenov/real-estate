package db

import (
	"context"
	"fmt"
	"github.com/begenov/real-estate/internal/config"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(ctx context.Context, cfg config.MinioConfig) (*minio.Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.Secure,
	})
	if err != nil {
		logger.Error(err)
		return nil, fmt.Errorf("ошибка подключения к MinIO: %w", err)
	}

	for _, bucket := range cfg.Buckets {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil {
			logger.Error(err)
			return nil, fmt.Errorf("ошибка проверки бакета %s: %w", bucket, err)
		}

		if !exists {
			err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				logger.Error(err)
				return nil, err
			}
		}

		policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [
				{
					"Effect": "Allow",
					"Principal": "*",
					"Action": ["s3:GetObject"],
					"Resource": ["arn:aws:s3:::%s/*"]
				}
			]
		}`, bucket)

		err = client.SetBucketPolicy(ctx, bucket, policy)
		if err != nil {
			logger.Error(err)
			return nil, err
		}
	}

	return client, nil
}
