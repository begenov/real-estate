package service

import (
	"context"
	"fmt"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
)

type IPageService interface {
	GetPage(ctx context.Context, id int64) (*model.Page, error)
	GetPages(ctx context.Context) ([]*model.Page, error)
	Create(ctx context.Context, page *model.Page) error
	Update(ctx context.Context, page *model.Page) error
	Delete(ctx context.Context, id int64) error
}

type PageService struct {
	pageRepo postgres.IPageRepo
}

func NewPageService(pageRepo postgres.IPageRepo) *PageService {
	return &PageService{
		pageRepo: pageRepo,
	}
}

func (p *PageService) GetPage(ctx context.Context, id int64) (*model.Page, error) {
	return p.pageRepo.GetPage(ctx, id)
}

func (p *PageService) GetPages(ctx context.Context) ([]*model.Page, error) {
	return p.pageRepo.GetPages(ctx)
}

func (p *PageService) Create(ctx context.Context, page *model.Page) error {
	for code, tr := range page.Translation {
		langID, ok := model.LanguageMap[code]
		if !ok {
			return fmt.Errorf("unknown language code: %s", code)
		}

		tr.Language.ID = langID
		tr.Language.Code = code
		page.Translation[code] = tr
	}

	return p.pageRepo.Create(ctx, page)
}

func (p *PageService) Update(ctx context.Context, page *model.Page) error {
	for code, tr := range page.Translation {
		langID, ok := model.LanguageMap[code]
		if !ok {
			return fmt.Errorf("unknown language code: %s", code)
		}

		tr.Language.ID = langID
		tr.Language.Code = code
		page.Translation[code] = tr
	}

	return p.pageRepo.Update(ctx, page)
}

func (p *PageService) Delete(ctx context.Context, id int64) error {
	return p.pageRepo.Delete(ctx, id)
}
