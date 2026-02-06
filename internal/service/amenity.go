package service

import (
	"context"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
)

type IAmenityService interface {
	Create(ctx context.Context, amenity *model.Amenity) error
	Update(ctx context.Context, amenity *model.Amenity) error
	Delete(ctx context.Context, amenityId int64) error
	GetAll(ctx context.Context) ([]*model.Amenity, error)
	GetByID(ctx context.Context, id int64) (*model.Amenity, error)
}

type AmenityService struct {
	amenityRepo postgres.IAmenityRepo
}

func NewAmenityService(amenityRepo postgres.IAmenityRepo) *AmenityService {
	return &AmenityService{amenityRepo: amenityRepo}
}

func (s *AmenityService) Create(ctx context.Context, amenity *model.Amenity) error {
	return s.amenityRepo.Create(ctx, amenity)
}

func (s *AmenityService) Update(ctx context.Context, amenity *model.Amenity) error {
	return s.amenityRepo.Update(ctx, amenity)
}

func (s *AmenityService) Delete(ctx context.Context, amenityId int64) error {
	return s.amenityRepo.Delete(ctx, amenityId)
}

func (s *AmenityService) GetAll(ctx context.Context) ([]*model.Amenity, error) {
	return s.amenityRepo.GetAll(ctx)
}

func (s *AmenityService) GetByID(ctx context.Context, id int64) (*model.Amenity, error) {
	return s.amenityRepo.GetByID(ctx, id)
}
