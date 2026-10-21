package service

import (
	"context"

	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
)

type ITourService interface {
	Submit(ctx context.Context, input *model.TourRequestInput) (*model.TourRequest, error)
	GetRequests(ctx context.Context, filter model.TourRequestFilter) ([]model.TourRequest, int, error)
	GetByID(ctx context.Context, id int64) (*model.TourRequest, error)
	UpdateStatus(ctx context.Context, id int64, statusID int) error
}

type TourService struct {
	tourRepo     postgres.ITourRepo
	emailService IEmailService
}

func NewTourService(tourRepo postgres.ITourRepo, emailService IEmailService) ITourService {
	return &TourService{
		tourRepo:     tourRepo,
		emailService: emailService,
	}
}

func (s *TourService) Submit(ctx context.Context, input *model.TourRequestInput) (*model.TourRequest, error) {
	req := &model.TourRequest{
		Name:         input.Name,
		Phone:        input.Phone,
		DesiredDate:  input.DesiredDate,
		RealEstateID: input.RealEstateID,
		StatusID:     model.TourStatusNew,
	}

	if err := s.tourRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	_ = s.emailService.SendTourForm(ctx, input.Name, input.Phone, input.DesiredDate)

	return req, nil
}

func (s *TourService) GetRequests(ctx context.Context, filter model.TourRequestFilter) ([]model.TourRequest, int, error) {
	return s.tourRepo.GetAll(ctx, filter)
}

func (s *TourService) GetByID(ctx context.Context, id int64) (*model.TourRequest, error) {
	return s.tourRepo.GetByID(ctx, id)
}

func (s *TourService) UpdateStatus(ctx context.Context, id int64, statusID int) error {
	switch statusID {
	case model.TourStatusNew, model.TourStatusInProgress, model.TourStatusCompleted, model.TourStatusCancelled:
	default:
		return model.ErrInvalidStatus
	}

	if _, err := s.tourRepo.GetByID(ctx, id); err != nil {
		return err
	}

	return s.tourRepo.UpdateStatus(ctx, id, statusID)
}
