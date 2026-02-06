package service

import (
	"context"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
)

type ILocationService interface {
	// Country methods
	GetCountries(ctx context.Context) ([]*model.Country, error)
	CreateCountry(ctx context.Context, country *model.Country) error
	UpdateCountry(ctx context.Context, country *model.Country) error
	DeleteCountry(ctx context.Context, countryId int64) error
	GetCountryByID(ctx context.Context, id int64) (*model.Country, error)

	// Region methods
	GetRegions(ctx context.Context, filter *model.RegionFilter) ([]*model.Region, int, error)
	CreateRegion(ctx context.Context, region *model.Region) error
	UpdateRegion(ctx context.Context, region *model.Region) error
	DeleteRegion(ctx context.Context, regionId int64) error
	GetRegionByID(ctx context.Context, id int64) (*model.Region, error)

	// District methods
	GetDistricts(ctx context.Context, filter *model.DistrictFilter) ([]*model.District, int, error)
	CreateDistrict(ctx context.Context, district *model.District) error
	UpdateDistrict(ctx context.Context, district *model.District) error
	DeleteDistrict(ctx context.Context, districtId int64) error
	GetDistrictByID(ctx context.Context, id int64) (*model.District, error)
}
type LocationService struct {
	locationRepo postgres.ILocationRepo
}

func NewLocationService(locationRepo postgres.ILocationRepo) *LocationService {
	return &LocationService{locationRepo: locationRepo}
}

// Country methods

func (s *LocationService) GetCountries(ctx context.Context) ([]*model.Country, error) {
	return s.locationRepo.GetCountries(ctx)
}

func (s *LocationService) CreateCountry(ctx context.Context, country *model.Country) error {
	return s.locationRepo.CreateCountry(ctx, country)
}

func (s *LocationService) UpdateCountry(ctx context.Context, country *model.Country) error {
	return s.locationRepo.UpdateCountry(ctx, country)
}

func (s *LocationService) DeleteCountry(ctx context.Context, countryId int64) error {
	return s.locationRepo.DeleteCountry(ctx, countryId)
}

func (s *LocationService) GetCountryByID(ctx context.Context, id int64) (*model.Country, error) {
	return s.locationRepo.GetCountryByID(ctx, id)
}

// Region methods

func (s *LocationService) GetRegions(ctx context.Context, filter *model.RegionFilter) ([]*model.Region, int, error) {
	regions, total, err := s.locationRepo.GetRegions(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	for i := range regions {
		regions[i].Country, err = s.locationRepo.GetCountryByID(ctx, regions[i].Country.ID)
		if err != nil {
			return nil, 0, err
		}
	}

	return regions, total, nil
}

func (s *LocationService) CreateRegion(ctx context.Context, region *model.Region) error {

	if region.Country == nil {
		logger.Error("Empty country")
		return model.ErrDataIsEmpty
	}

	return s.locationRepo.CreateRegion(ctx, region)
}

func (s *LocationService) UpdateRegion(ctx context.Context, region *model.Region) error {
	if region.Country == nil {
		logger.Error("Empty country")
		return model.ErrDataIsEmpty
	}

	return s.locationRepo.UpdateRegion(ctx, region)
}

func (s *LocationService) DeleteRegion(ctx context.Context, regionId int64) error {
	return s.locationRepo.DeleteRegion(ctx, regionId)
}

func (s *LocationService) GetRegionByID(ctx context.Context, id int64) (*model.Region, error) {
	region, err := s.locationRepo.GetRegionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	region.Country, err = s.locationRepo.GetCountryByID(ctx, region.Country.ID)
	if err != nil {
		return nil, err
	}

	return region, nil
}

// District methods

func (s *LocationService) GetDistricts(ctx context.Context, filter *model.DistrictFilter) ([]*model.District, int, error) {
	districts, total, err := s.locationRepo.GetDistricts(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	for i := range districts {
		districts[i].Region, err = s.locationRepo.GetRegionByID(ctx, districts[i].Region.ID)
		if err != nil {
			return nil, 0, err
		}
	}

	return districts, total, nil
}

func (s *LocationService) CreateDistrict(ctx context.Context, district *model.District) error {
	if district.Region == nil {
		logger.Error("Empty region")
		return model.ErrDataIsEmpty
	}

	return s.locationRepo.CreateDistrict(ctx, district)
}

func (s *LocationService) UpdateDistrict(ctx context.Context, district *model.District) error {
	if district.Region == nil {
		logger.Error("Empty region")
		return model.ErrDataIsEmpty
	}

	return s.locationRepo.UpdateDistrict(ctx, district)
}

func (s *LocationService) DeleteDistrict(ctx context.Context, districtId int64) error {
	return s.locationRepo.DeleteDistrict(ctx, districtId)
}

func (s *LocationService) GetDistrictByID(ctx context.Context, id int64) (*model.District, error) {
	district, err := s.locationRepo.GetDistrictByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if district == nil {
		return nil, nil
	}

	district.Region, err = s.locationRepo.GetRegionByID(ctx, district.Region.ID)
	if err != nil {
		return nil, err
	}

	return district, nil
}
