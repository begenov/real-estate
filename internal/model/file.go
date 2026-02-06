package model

import "io"

type (
	FileType string
)

const (
	Image FileType = "image"
)

const (
	Estate  = "real-estate"
	Avatars = "users-avatars"
)

type UploadInput struct {
	File        io.Reader
	Size        int64
	Name        string
	ContentType string
	Bucket      string
	FileType    FileType
	UserID      int64
	NoWatermark bool
}

type File struct {
	Id  int64  `json:"id"`
	URL string `json:"url"`
}
