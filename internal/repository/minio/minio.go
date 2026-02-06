package minio

import (
	"context"
	"github.com/begenov/real-estate/internal/model"
	"github.com/minio/minio-go/v7"
)

type IMinioRepo interface {
	Upload(ctx context.Context, req *model.UploadInput) error
	Delete(ctx context.Context, bucket, objectName string) error
	Get(ctx context.Context, bucket, objectName string) (*minio.Object, minio.ObjectInfo, error)
}

type MinioRepo struct {
	client *minio.Client
}

func NewMinioRepo(minio *minio.Client) *MinioRepo {
	var minioRepo MinioRepo

	minioRepo.client = minio

	return &minioRepo
}

func (m *MinioRepo) Upload(ctx context.Context, input *model.UploadInput) error {
	opts := minio.PutObjectOptions{
		ContentType:  input.ContentType,
		UserMetadata: map[string]string{"x-amz-acl": "public-read"},
	}

	_, err := m.client.PutObject(ctx, input.Bucket, input.Name, input.File, input.Size, opts)
	if err != nil {
		return err
	}

	return nil
}

func (m *MinioRepo) Delete(ctx context.Context, bucket, objectName string) error {
	return m.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func (m *MinioRepo) Get(ctx context.Context, bucket, objectName string) (*minio.Object, minio.ObjectInfo, error) {
	obj, err := m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, minio.ObjectInfo{}, err
	}

	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, minio.ObjectInfo{}, err
	}

	return obj, info, nil
}
