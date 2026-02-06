package translate

import (
	"context"

	"cloud.google.com/go/translate"
	"google.golang.org/api/option"
)

func NewTranslateClient(ctx context.Context, path string) (*translate.Client, error) {
	var (
		service *translate.Client
		err     error
	)
	if path == "" {
		service, err = translate.NewClient(ctx)
	} else {
		service, err = translate.NewClient(ctx, option.WithCredentialsFile(path))
	}
	if err != nil {
		return nil, err
	}

	return service, nil
}
