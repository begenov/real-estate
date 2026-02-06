package service

import (
	"bytes"
	"context"
	"github.com/begenov/real-estate/internal/async"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/minio"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

var folders = map[model.FileType]string{
	model.Image: "images",
}

type MinioService struct {
	fileRepo  postgres.IFileRepo
	minioRepo minio.IMinioRepo
	buckets   map[string]struct{}
	endpoint  string
	env       string

	imageService     IImageService
	reprocessWorkers int
	reprocessQueue   int
}

type IMinioService interface {
	Upload(ctx context.Context, input *model.UploadInput) (*model.File, error)
	UploadUserPhoto(ctx context.Context, input *model.UploadInput) (string, error)
	DeleteByURLs(ctx context.Context, urls []string) error
}

func NewMinioService(fileRepo postgres.IFileRepo, minioRepo minio.IMinioRepo, imageService IImageService, endpoint, env string, reprocessWorkers, reprocessQueue int, buckets ...string) *MinioService {
	var minioService = MinioService{
		fileRepo:         fileRepo,
		minioRepo:        minioRepo,
		endpoint:         endpoint,
		buckets:          make(map[string]struct{}, len(buckets)),
		env:              env,
		imageService:     imageService,
		reprocessWorkers: reprocessWorkers,
		reprocessQueue:   reprocessQueue,
	}

	for i := range buckets {
		minioService.buckets[buckets[i]] = struct{}{}
	}

	return &minioService
}

func (s *MinioService) Upload(ctx context.Context, input *model.UploadInput) (*model.File, error) {
	if ok := s.checkBucket(input.Bucket); !ok {
		return nil, model.ErrBucketNotExist
	}

	if strings.HasPrefix(input.ContentType, "image/") {
		if err := s.processImage(input); err != nil {
			return nil, err
		}
	}

	input.Name = s.generateFilename(input)

	err := s.minioRepo.Upload(ctx, input)
	if err != nil {
		return nil, err
	}

	url := s.generateFileURL(input.Name, input.Bucket)

	id, err := s.fileRepo.Create(ctx, url, input.UserID)
	if err != nil {
		return nil, err
	}

	return &model.File{Id: id, URL: url}, nil
}

func (s *MinioService) UploadUserPhoto(ctx context.Context, input *model.UploadInput) (string, error) {
	input.NoWatermark = true

	if strings.HasPrefix(input.ContentType, "image/") {
		if err := s.processImage(input); err != nil {
			return "", err
		}
	}

	input.Name = s.generateFilename(input)

	err := s.minioRepo.Upload(ctx, input)
	if err != nil {
		return "", err
	}

	url := s.generateFileURL(input.Name, input.Bucket)

	return url, nil
}

func (s *MinioService) DeleteByURLs(ctx context.Context, urls []string) error {
	for _, url := range urls {
		if url == "" {
			continue
		}

		bucket, objectName, err := parseFileURL(url)
		if err != nil {
			return err
		}

		if ok := s.checkBucket(bucket); !ok {
			return model.ErrBucketNotExist
		}

		if err := s.minioRepo.Delete(ctx, bucket, objectName); err != nil {
			return err
		}
	}

	return nil
}

func (s *MinioService) ReprocessExistingImages(ctx context.Context, batchSize int, pause time.Duration) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	workers := s.reprocessWorkers
	if workers <= 0 {
		workers = 4
	}
	queue := s.reprocessQueue
	if queue <= 0 {
		queue = batchSize
	}
	pool := async.NewWorkerPool(workers, queue)
	pool.Start(ctx, workers)
	defer func() {
		pool.Close()
		pool.Wait()
	}()

	offset := 0
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		urls, err := s.fileRepo.ListURLs(ctx, batchSize, offset)
		if err != nil {
			return err
		}
		if len(urls) == 0 {
			return nil
		}

		var wg sync.WaitGroup
		for _, url := range urls {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			currentURL := url
			wg.Add(1)
			if err := pool.Submit(ctx, async.WithTiming("reprocess_image", func(ctx context.Context) error {
				defer wg.Done()
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if err := s.reprocessImage(ctx, currentURL); err != nil {
					logger.Warn("reprocess image error: ", err)
				}
				return nil
			})); err != nil {
				wg.Done()
				return err
			}
		}
		wg.Wait()

		offset += len(urls)

		logger.Info("Reprocessed images: ", offset)
		if pause > 0 {
			time.Sleep(pause)
		}
	}
}

func (s *MinioService) reprocessImage(ctx context.Context, url string) error {
	if url == "" {
		return nil
	}

	bucket, objectName, err := parseFileURL(url)
	if err != nil {
		return err
	}

	if ok := s.checkBucket(bucket); !ok {
		return model.ErrBucketNotExist
	}

	obj, info, err := s.minioRepo.Get(ctx, bucket, objectName)
	if err != nil {
		return err
	}
	defer func() {
		_ = obj.Close()
	}()

	content, err := io.ReadAll(obj)
	if err != nil {
		return err
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = http.DetectContentType(content)
	}

	if contentType != "image/png" && contentType != "image/jpeg" && contentType != "image/jpg" {
		return nil
	}

	compressedReader, newContentType, err := s.imageService.Compress(bytes.NewReader(content), contentType)
	if err != nil {
		return err
	}

	compressed, err := io.ReadAll(compressedReader)
	if err != nil {
		return err
	}

	input := &model.UploadInput{
		File:        bytes.NewReader(compressed),
		Size:        int64(len(compressed)),
		Name:        objectName,
		ContentType: newContentType,
		Bucket:      bucket,
		FileType:    model.Image,
	}

	return s.minioRepo.Upload(ctx, input)
}

func (s *MinioService) processImage(input *model.UploadInput) error {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, input.File); err != nil {
		return err
	}

	content := buf.Bytes()
	if !input.NoWatermark {
		modifiedReader, err := s.imageService.AddWatermark(bytes.NewReader(content))
		if err != nil {
			return err
		}

		modifiedBytes, err := io.ReadAll(modifiedReader)
		if err != nil {
			return err
		}

		content = modifiedBytes
		input.ContentType = "image/png"
	}

	if input.ContentType == "image/png" || input.ContentType == "image/jpeg" || input.ContentType == "image/jpg" {
		compressedReader, newContentType, err := s.imageService.Compress(bytes.NewReader(content), input.ContentType)
		if err != nil {
			return err
		}

		compressedBytes, err := io.ReadAll(compressedReader)
		if err != nil {
			return err
		}

		content = compressedBytes
		input.ContentType = newContentType
	}

	input.File = bytes.NewReader(content)
	input.Size = int64(len(content))

	return nil
}
