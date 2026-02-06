package service

import (
	"fmt"
	"github.com/begenov/real-estate/internal/model"
	"github.com/google/uuid"
	"strings"
)

func (s *MinioService) checkBucket(bucket string) bool {
	_, ok := s.buckets[bucket]
	return ok
}

func (s *MinioService) generateFilename(file *model.UploadInput) string {
	filename := fmt.Sprintf("%s.%s", uuid.New().String(), getFileExtension(file.Name))
	folder := folders[file.FileType]

	return fmt.Sprintf("%s/%s/%s", s.env, folder, filename)
}

func getFileExtension(filename string) string {
	parts := strings.Split(filename, ".")

	return parts[len(parts)-1]
}

func (s *MinioService) generateFileURL(filename, bucket string) string {
	return fmt.Sprintf("%s/%s", bucket, filename)
}

func parseFileURL(url string) (string, string, error) {
	parts := strings.SplitN(url, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid file url: %s", url)
	}

	return parts[0], parts[1], nil
}
