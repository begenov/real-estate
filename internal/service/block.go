package service

import (
	"context"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
)

type IBlockService interface {
	Create(ctx context.Context, block *model.Block) error
	Update(ctx context.Context, block *model.Block) error
	GetBlock(ctx context.Context, id int64) (*model.Block, error)
	GetBlocks(ctx context.Context, pageID int64) ([]*model.Block, error)
	Delete(ctx context.Context, id int64) error
}

type BlockService struct {
	blockRepo        *postgres.BlockRepo
	translateService ITranslateService
}

func NewBlockService(blockRepo *postgres.BlockRepo, translateService ITranslateService) *BlockService {
	return &BlockService{
		blockRepo:        blockRepo,
		translateService: translateService,
	}
}

func (s *BlockService) Create(ctx context.Context, block *model.Block) error {
	return s.blockRepo.Create(ctx, block)
}

func (s *BlockService) Update(ctx context.Context, block *model.Block) error {
	targetLangs := []string{"en", "de", "tr"}

	for i := range block.BlockContents {
		content := block.BlockContents[i]

		if content.Translation == nil {
			continue
		}

		baseText := content.Translation["ru"].Description
		baseTitle := content.Translation["ru"].Name

		for _, lang := range targetLangs {
			titleTr, err := s.translateService.Translate(ctx, baseTitle, lang)
			if err != nil {
				logger.Error("s.translateService.Translate: %w", err)
				continue
			}

			bodyTr, err := s.translateService.Translate(ctx, baseText, lang)
			if err != nil {
				logger.Error("s.translateService.Translate: %w", err)
				continue
			}

			content.Translation[lang] = model.BlockTranslation{
				Name:        titleTr,
				Description: bodyTr,
			}
		}
	}

	return s.blockRepo.Update(ctx, block)
}

func (s *BlockService) GetBlock(ctx context.Context, id int64) (*model.Block, error) {
	return s.blockRepo.GetBlock(ctx, id)
}

func (s *BlockService) GetBlocks(ctx context.Context, pageID int64) ([]*model.Block, error) {
	return s.blockRepo.GetBlocks(ctx, pageID)
}

func (s *BlockService) Delete(ctx context.Context, id int64) error {
	return s.blockRepo.Delete(ctx, id)
}
