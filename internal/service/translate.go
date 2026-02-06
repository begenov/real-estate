package service

import (
	"context"
	"fmt"

	"cloud.google.com/go/translate"
	"golang.org/x/text/language"
)

type ITranslateService interface {
	Translate(ctx context.Context, text string, targetLang string) (string, error)
}

type TranslateService struct {
	client *translate.Client
}

func NewTranslateService(client *translate.Client) *TranslateService {
	return &TranslateService{
		client: client,
	}
}

func (s *TranslateService) TranslateText(ctx context.Context, text string, targets []string) (map[string]string, error) {
	results := make(map[string]string)

	for _, langCode := range targets {
		lang, err := language.Parse(langCode)
		if err != nil {
			return nil, fmt.Errorf("ошибка парсинга языка %s: %w", langCode, err)
		}

		resp, err := s.client.Translate(ctx, []string{text}, lang, nil)
		if err != nil {
			return nil, fmt.Errorf("ошибка перевода на %s: %w", langCode, err)
		}

		if len(resp) > 0 {
			results[langCode] = resp[0].Text
		}
	}

	return results, nil
}

func (s *TranslateService) Translate(ctx context.Context, text string, targetLang string) (string, error) {
	lang, err := language.Parse(targetLang)
	if err != nil {
		return "", fmt.Errorf("ошибка парсинга языка %s: %w", targetLang, err)
	}

	resp, err := s.client.Translate(ctx, []string{text}, lang, nil)
	if err != nil {
		return "", fmt.Errorf("ошибка перевода на %s: %w", targetLang, err)
	}

	if len(resp) > 0 {
		return resp[0].Text, nil
	}

	return "", fmt.Errorf("перевод на %s не найден", targetLang)
}
