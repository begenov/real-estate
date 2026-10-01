package service

import (
	"context"
	"errors"
	"testing"

	"github.com/begenov/real-estate/internal/model"
)

func defaultRates() IExchangeRateService {
	return &mockExchangeRate{
		GetRatesFn: func(_ context.Context) (map[string]float64, error) {
			return map[string]float64{"USD": 1.2, "EUR": 1.1}, nil
		},
	}
}

func defaultLocation() *mockLocationService {
	return &mockLocationService{
		GetRegionByIDFn: func(_ context.Context, id int64) (*model.Region, error) {
			return &model.Region{ID: id}, nil
		},
		GetDistrictByIDFn: func(_ context.Context, id int64) (*model.District, error) {
			return &model.District{ID: id}, nil
		},
	}
}

func defaultAmenity() *mockAmenityService {
	return &mockAmenityService{
		GetByIDsFn: func(_ context.Context, _ []int64) ([]*model.Amenity, error) {
			return nil, nil
		},
	}
}

// ── GetRealEstate ─────────────────────────────────────────────────────────────

func TestGetRealEstate_InvalidID(t *testing.T) {
	svc := NewRealEstateService(
		&mockRealEstateRepo{}, &mockUserRepo{}, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	for _, id := range []int64{0, -1, -100} {
		_, err := svc.GetRealEstate(context.Background(), id, nil)
		if !errors.Is(err, model.ErrIdMustBeGreaterThanZero) {
			t.Fatalf("id=%d: want ErrIdMustBeGreaterThanZero, got %v", id, err)
		}
	}
}

func TestGetRealEstate_NotFound(t *testing.T) {
	reRepo := &mockRealEstateRepo{
		GetRealEstateFn: func(_ context.Context, _ int64) (*model.RealEstate, error) {
			return nil, model.ErrNotFound
		},
	}

	svc := NewRealEstateService(
		reRepo, &mockUserRepo{}, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	_, err := svc.GetRealEstate(context.Background(), 1, nil)
	if !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestGetRealEstate_Success(t *testing.T) {
	estate := &model.RealEstate{
		ID:       42,
		Price:    100_000,
		Region:   model.Region{ID: 1},
		District: model.District{ID: 2},
	}

	reRepo := &mockRealEstateRepo{
		GetRealEstateFn: func(_ context.Context, _ int64) (*model.RealEstate, error) {
			return estate, nil
		},
		GetRealEstateURLsFn: func(_ context.Context, _ int64) ([]model.RealEstatePhoto, error) {
			return nil, nil
		},
		GetRealEstateLastHistoryFn: func(_ context.Context, _ int64) (*model.RealEstateHistory, error) {
			return &model.RealEstateHistory{Status: model.Status{ID: model.Status_Available}}, nil
		},
		GetRealEstateTranslationsFn: func(_ context.Context, _ int64) ([]model.RealEstateTranslation, error) {
			return nil, nil
		},
		GetAmenitiesByRealEstateIDFn: func(_ context.Context, _ int64) ([]int64, error) {
			return nil, nil
		},
	}

	svc := NewRealEstateService(
		reRepo, &mockUserRepo{}, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	result, err := svc.GetRealEstate(context.Background(), 42, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 42 {
		t.Fatalf("want ID=42, got %d", result.ID)
	}
	if result.SerialNumber != "000000042" {
		t.Fatalf("want SerialNumber=000000042, got %s", result.SerialNumber)
	}
	if result.PriceUSD == 0 {
		t.Fatal("expected non-zero PriceUSD from exchange rates")
	}
}

// ── CreateRealEstate ──────────────────────────────────────────────────────────

func TestCreateRealEstate_NilInput(t *testing.T) {
	svc := NewRealEstateService(
		&mockRealEstateRepo{}, &mockUserRepo{}, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	err := svc.CreateRealEstate(context.Background(), nil)
	if !errors.Is(err, model.ErrDataIsEmpty) {
		t.Fatalf("want ErrDataIsEmpty, got %v", err)
	}
}

func TestCreateRealEstate_AccessDenied(t *testing.T) {
	userRepo := &mockUserRepo{
		GetUserRoleFn: func(_ context.Context, _ int64) ([]model.Role, error) {
			return []model.Role{}, nil
		},
	}

	svc := NewRealEstateService(
		&mockRealEstateRepo{}, userRepo, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	err := svc.CreateRealEstate(context.Background(), validRealEstateInput())
	if !errors.Is(err, model.ErrAccessDenied) {
		t.Fatalf("want ErrAccessDenied, got %v", err)
	}
}

// ── DeleteRealEstate ──────────────────────────────────────────────────────────

func TestDeleteRealEstate_InvalidID(t *testing.T) {
	svc := NewRealEstateService(
		&mockRealEstateRepo{}, &mockUserRepo{}, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	err := svc.DeleteRealEstate(context.Background(), 0, 1)
	if !errors.Is(err, model.ErrIdMustBeGreaterThanZero) {
		t.Fatalf("want ErrIdMustBeGreaterThanZero, got %v", err)
	}
}

func TestDeleteRealEstate_AccessDenied(t *testing.T) {
	reRepo := &mockRealEstateRepo{
		GetRealEstateFn: func(_ context.Context, _ int64) (*model.RealEstate, error) {
			return &model.RealEstate{ID: 1, Manager: model.User{ID: 10}}, nil
		},
	}
	userRepo := &mockUserRepo{
		GetUserRoleFn: func(_ context.Context, _ int64) ([]model.Role, error) {
			return []model.Role{{Code: model.Role_Manager}}, nil
		},
	}

	svc := NewRealEstateService(
		reRepo, userRepo, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	err := svc.DeleteRealEstate(context.Background(), 1, 99)
	if !errors.Is(err, model.ErrAccessDenied) {
		t.Fatalf("want ErrAccessDenied, got %v", err)
	}
}

// ── UpdateStatus ──────────────────────────────────────────────────────────────

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	svc := NewRealEstateService(
		&mockRealEstateRepo{}, &mockUserRepo{}, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	err := svc.UpdateStatus(context.Background(), 1, 1, 999)
	if !errors.Is(err, model.ErrRealEstateStatus) {
		t.Fatalf("want ErrRealEstateStatus, got %v", err)
	}
}

func TestUpdateStatus_SameStatus_NoOp(t *testing.T) {
	reRepo := &mockRealEstateRepo{
		GetRealEstateFn: func(_ context.Context, _ int64) (*model.RealEstate, error) {
			return &model.RealEstate{ID: 1, Manager: model.User{ID: 1}}, nil
		},
		GetRealEstateLastHistoryFn: func(_ context.Context, _ int64) (*model.RealEstateHistory, error) {
			return &model.RealEstateHistory{Status: model.Status{ID: model.Status_Available}}, nil
		},
	}
	userRepo := &mockUserRepo{
		GetUserRoleFn: func(_ context.Context, _ int64) ([]model.Role, error) {
			return []model.Role{{Code: model.Role_Admin}}, nil
		},
	}

	svc := NewRealEstateService(
		reRepo, userRepo, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	err := svc.UpdateStatus(context.Background(), 1, 1, model.Status_Available)
	if err != nil {
		t.Fatalf("unexpected error on same-status update: %v", err)
	}
}

func TestUpdateStatus_ManagerCannotSetAvailable(t *testing.T) {
	const managerID = int64(5)
	reRepo := &mockRealEstateRepo{
		GetRealEstateFn: func(_ context.Context, _ int64) (*model.RealEstate, error) {
			return &model.RealEstate{ID: 1, Manager: model.User{ID: managerID}}, nil
		},
	}
	userRepo := &mockUserRepo{
		GetUserRoleFn: func(_ context.Context, _ int64) ([]model.Role, error) {
			return []model.Role{{Code: model.Role_Manager}}, nil
		},
	}

	svc := NewRealEstateService(
		reRepo, userRepo, &mockTxRepo{}, &mockMinioService{},
		defaultAmenity(), defaultLocation(), defaultRates(), &mockTranslate{}, 1,
	)

	err := svc.UpdateStatus(context.Background(), 1, managerID, model.Status_Available)
	if !errors.Is(err, model.ErrAccessDenied) {
		t.Fatalf("want ErrAccessDenied, got %v", err)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func validRealEstateInput() *model.RealEstateInput {
	return &model.RealEstateInput{
		OwnerId:   1,
		Price:     100_000,
		Area:      60,
		Latitude:  40.0,
		Longitude: 50.0,
		RegionID:  1,
		Type:      model.Type{ID: 1},
		Purpose:   model.RealEstatePurpose{ID: 1},
		Translations: []model.RealEstateTranslationInput{
			{Title: "Test", Description: "Desc", Language: model.Language{Code: "ru", ID: 1}},
		},
		NewPhotos: []model.RealEstatePhoto{{ID: 1, URL: "http://example.com/photo.jpg"}},
	}
}
