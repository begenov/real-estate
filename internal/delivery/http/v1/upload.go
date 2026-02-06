package v1

import (
	"bytes"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	maxUploadSize = 5 << 20 // 5 megabytes
)

var (
	imageTypes = map[string]struct{}{
		"image/jpeg": struct{}{},
		"image/png":  struct{}{},
	}
)

func (h *Handler) initUploadRoutes(api *gin.RouterGroup) {
	upload := api.Group("/upload")
	upload.Use(h.userIdentity)
	{
		upload.POST("/image", h.uploadImage)
	}
}

func (h *Handler) uploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)

	withoutWM := c.DefaultQuery("without_watermark", c.PostForm("without_watermark"))
	noWatermark := withoutWM == "true"

	file, fileHeader, err := c.Request.FormFile("file")
	if err != nil {
		logger.Error(err)
		_ = c.Error(err)
		return
	}

	defer func() {
		_ = file.Close()
	}()

	var buffer bytes.Buffer
	if _, err := io.Copy(&buffer, file); err != nil {
		logger.Error(err)
		_ = c.Error(err)
		return
	}

	contentType := http.DetectContentType(buffer.Bytes())
	if _, ex := imageTypes[contentType]; !ex {
		logger.Error(err)
		_ = c.Error(model.ErrFileNotSupported)
		return
	}

	id, err := getUserId(c)
	if err != nil {
		_ = c.Error(err)
		return
	}

	var inp = model.UploadInput{
		File:        bytes.NewReader(buffer.Bytes()),
		Name:        fileHeader.Filename,
		Size:        fileHeader.Size,
		ContentType: contentType,
		Bucket:      model.Estate,
		FileType:    model.Image,
		UserID:      id,
		NoWatermark: noWatermark,
	}

	res, err := h.minioService.Upload(c.Request.Context(), &inp)
	if err != nil {
		logger.Error(err)
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, res)
}
