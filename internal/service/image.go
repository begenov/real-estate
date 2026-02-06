package service

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"

	"github.com/disintegration/imaging"
)

type IImageService interface {
	AddWatermark(input io.Reader) (io.Reader, error)
	Compress(input io.Reader, contentType string) (io.Reader, string, error)
}

type ImageService struct {
	path string
}

func NewImageService(path string) *ImageService {
	return &ImageService{
		path: path,
	}
}

func (s *ImageService) AddWatermark(input io.Reader) (io.Reader, error) {
	img, err := imaging.Decode(input)
	if err != nil {
		return nil, err
	}

	watermarkFile, err := os.Open(s.path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = watermarkFile.Close()
	}()

	watermark, err := imaging.Decode(watermarkFile)
	if err != nil {
		return nil, err
	}

	imgW := img.Bounds().Dx()
	imgH := img.Bounds().Dy()

	maxWMWidth := imgW / 5
	if maxWMWidth < 1 {
		maxWMWidth = 1
	}

	watermark = imaging.Resize(watermark, maxWMWidth, 0, imaging.Lanczos)

	wmW := watermark.Bounds().Dx()
	wmH := watermark.Bounds().Dy()

	x := 10
	y := 10

	if x+wmW > imgW {
		x = imgW - wmW
	}
	if y+wmH > imgH {
		y = imgH - wmH
	}
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}

	position := image.Pt(x, y)

	result := imaging.Clone(img)
	result = imaging.Overlay(result, watermark, position, 0.7)

	buf := new(bytes.Buffer)
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	err = encoder.Encode(buf, result)
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (s *ImageService) Compress(input io.Reader, contentType string) (io.Reader, string, error) {
	img, err := imaging.Decode(input)
	if err != nil {
		return nil, contentType, err
	}

	buf := new(bytes.Buffer)
	switch contentType {
	case "image/png":
		encoder := png.Encoder{CompressionLevel: png.BestCompression}
		if err := encoder.Encode(buf, img); err != nil {
			return nil, contentType, err
		}
		return buf, "image/png", nil
	case "image/jpeg", "image/jpg":
		if err := jpeg.Encode(buf, img, &jpeg.Options{Quality: 75}); err != nil {
			return nil, contentType, err
		}
		return buf, "image/jpeg", nil
	default:
		return nil, contentType, fmt.Errorf("unsupported image content type: %s", contentType)
	}
}
