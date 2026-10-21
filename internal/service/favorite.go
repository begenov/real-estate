package service

import (
	"context"
	"fmt"

	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
)

type IFavoriteService interface {
	Add(ctx context.Context, userID, realEstateID int64) error
	Remove(ctx context.Context, userID, realEstateID int64) error
	GetFavorites(ctx context.Context, userID int64, page, rows int) ([]model.Favorite, int, error)
}

type FavoriteService struct {
	favoriteRepo   postgres.IFavoriteRepo
	realEstateRepo postgres.IRealEstateRepo
}

func NewFavoriteService(favoriteRepo postgres.IFavoriteRepo, realEstateRepo postgres.IRealEstateRepo) IFavoriteService {
	return &FavoriteService{
		favoriteRepo:   favoriteRepo,
		realEstateRepo: realEstateRepo,
	}
}

func (s *FavoriteService) Add(ctx context.Context, userID, realEstateID int64) error {
	estate, err := s.realEstateRepo.GetRealEstate(ctx, realEstateID)
	if err != nil || estate == nil {
		return model.ErrNotFound
	}

	exists, err := s.favoriteRepo.Exists(ctx, userID, realEstateID)
	if err != nil {
		return fmt.Errorf("check favorite exists: %w", err)
	}
	if exists {
		return nil
	}

	return s.favoriteRepo.Add(ctx, userID, realEstateID)
}

func (s *FavoriteService) Remove(ctx context.Context, userID, realEstateID int64) error {
	return s.favoriteRepo.Remove(ctx, userID, realEstateID)
}

func (s *FavoriteService) GetFavorites(ctx context.Context, userID int64, page, rows int) ([]model.Favorite, int, error) {
	favorites, total, err := s.favoriteRepo.GetByUserID(ctx, userID, page, rows)
	if err != nil {
		return nil, 0, err
	}

	for i := range favorites {
		estate, err := s.realEstateRepo.GetRealEstate(ctx, favorites[i].RealEstateID)
		if err != nil {
			logger.Errorf("GetRealEstate(%d): %v", favorites[i].RealEstateID, err)
			continue
		}

		urls, err := s.realEstateRepo.GetRealEstateURLs(ctx, estate.ID)
		if err != nil {
			logger.Errorf("GetRealEstateURLs(%d): %v", estate.ID, err)
		} else {
			estate.Photos = urls
		}

		translations, err := s.realEstateRepo.GetRealEstateTranslations(ctx, estate.ID)
		if err != nil {
			logger.Errorf("GetRealEstateTranslations(%d): %v", estate.ID, err)
		} else {
			estate.Translations = translations
		}

		favorites[i].RealEstate = estate
	}

	return favorites, total, nil
}
