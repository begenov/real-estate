package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/async"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"sync"
	"time"
)

type IRealEstateService interface {
	CreateRealEstate(ctx context.Context, estate *model.RealEstateInput) error
	GetRealEstates(ctx context.Context, filter model.RealEstateFilter, userId *int64) (model.RealEstates, int, error)
	GetRealEstate(ctx context.Context, estateId int64, userId *int64) (*model.RealEstate, error)
	GetRealEstateTotal(ctx context.Context, filter model.RealEstateFilter, userId *int64) (int, error)

	UpdateRealEstate(ctx context.Context, estate *model.RealEstateInput) error
	DeleteRealEstate(ctx context.Context, id, userId int64) error
	UpdateStatus(ctx context.Context, estateId, userId int64, statusId int) error
}

type RealEstateService struct {
	realEstateRepo       postgres.IRealEstateRepo
	userRepo             postgres.IUserRepo
	txRepo               postgres.ITxRepo
	fileService          IMinioService
	amenityService       IAmenityService
	locationService      ILocationService
	exchangeRateService  IExchangeRateService
	translator           ITranslateService
	translateConcurrency int
}

func NewRealEstateService(realEstateRepo postgres.IRealEstateRepo, userRepo postgres.IUserRepo,
	txRepo postgres.ITxRepo, fileService IMinioService, amenityService IAmenityService,
	locationService ILocationService, exchangeRateService IExchangeRateService,
	translateService ITranslateService, translateConcurrency int) IRealEstateService {
	if translateConcurrency <= 0 {
		translateConcurrency = 3
	}
	return &RealEstateService{
		realEstateRepo:       realEstateRepo,
		userRepo:             userRepo,
		txRepo:               txRepo,
		fileService:          fileService,
		amenityService:       amenityService,
		locationService:      locationService,
		exchangeRateService:  exchangeRateService,
		translator:           translateService,
		translateConcurrency: translateConcurrency,
	}
}

func (s *RealEstateService) GetRealEstates(ctx context.Context, filter model.RealEstateFilter, userId *int64) (model.RealEstates, int, error) {
	if filter.StatusID != nil {
		filter.Status = append(filter.Status, int(*filter.StatusID))
	}

	if userId == nil && len(filter.Status) == 0 {
		filter.Status = append(filter.Status, model.Status_Created, model.Status_Available)
	}

	if err := filter.Validate(); err != nil {
		return nil, 0, err
	}

	rates, err := s.exchangeRateService.GetRates(ctx)
	if err != nil {
		logger.Error("GetRates(): ", err)
		rates = map[string]float64{"USD": 0, "EUR": 0}
	}

	estates, total, err := s.realEstateRepo.GetRealEstates(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	if len(estates) == 0 {
		return estates, total, nil
	}

	estateIDs := make([]int64, 0, len(estates))
	for i := range estates {
		estateIDs = append(estateIDs, estates[i].ID)
	}

	translationMap, err := s.realEstateRepo.GetRealEstateTranslationsByIDs(ctx, estateIDs)
	if err != nil {
		logger.Error("s.realEstateRepo.GetRealEstateTranslationsByIDs(): ", err)
		return nil, 0, err
	}

	var historyMap map[int64]*model.RealEstateHistory
	if userId != nil {
		historyMap, err = s.realEstateRepo.GetRealEstateLastHistories(ctx, estateIDs)
		if err != nil {
			logger.Warnf("GetRealEstateLastHistories error: %v", err)
		}
	}

	for i := range estates {
		estates[i].PriceUSD = estates[i].Price * rates["USD"]
		estates[i].PriceEUR = estates[i].Price * rates["EUR"]

		estates[i].SerialNumber = fmt.Sprintf("%09d", estates[i].ID)

		if translations, ok := translationMap[estates[i].ID]; ok {
			estates[i].Translations = translations
		}

		if userId != nil && historyMap != nil {
			if history, ok := historyMap[estates[i].ID]; ok {
				estates[i].LatestHistory = history
			}
		}
	}

	return estates, total, nil
}

func (s *RealEstateService) GetRealEstate(ctx context.Context, estateId int64, userId *int64) (*model.RealEstate, error) {
	if estateId <= 0 {
		return nil, model.ErrIdMustBeGreaterThanZero
	}

	estate, err := s.realEstateRepo.GetRealEstate(ctx, estateId)
	if err != nil {
		logger.Error("s.realEstateRepo.GetRealEstate(): ", err)
		return nil, err
	}

	estate.Photos, err = s.realEstateRepo.GetRealEstateURLs(ctx, estate.ID)
	if err != nil {
		logger.Error("s.realEstateRepo.GetRealEstateURLs(): ", err)
		return nil, err
	}

	latestHistory, err := s.realEstateRepo.GetRealEstateLastHistory(ctx, estate.ID)
	if err != nil {
		logger.Error("s.realEstateRepo.GetRealEstateLastHistory(): ", err)
		return nil, err
	}

	if userId != nil {
		estate.LatestHistory = latestHistory
	}

	translations, err := s.realEstateRepo.GetRealEstateTranslations(ctx, estateId)
	if err != nil {
		logger.Error("s.realEstateRepo.GetRealEstateTranslations(): ", err)
		return nil, err
	}

	estate.Translations = translations

	amenityIDs, err := s.realEstateRepo.GetAmenitiesByRealEstateID(ctx, estate.ID)
	if err == nil && len(amenityIDs) > 0 {
		amenities, err := s.amenityService.GetByIDs(ctx, amenityIDs)
		if err != nil {
			return nil, err
		}
		estate.Amenities = amenities
	}

	region, err := s.locationService.GetRegionByID(ctx, estate.Region.ID)
	if err == nil && region != nil {
		estate.Region = *region
	} else {
		logger.Error("s.locationService.GetRegionByID(): ", err)
	}

	district, err := s.locationService.GetDistrictByID(ctx, estate.District.ID)
	if err == nil && district != nil {
		estate.District = *district
	} else {
		logger.Error("s.locationService.GetDistrictByID(): ", err)
	}

	rates, err := s.exchangeRateService.GetRates(ctx)
	if err != nil {
		logger.Error("getRatesGBPtoUSDandEUR(): ", err)
	} else {
		estate.PriceUSD = estate.Price * rates["USD"]
		estate.PriceEUR = estate.Price * rates["EUR"]
	}

	estate.SerialNumber = fmt.Sprintf("%09d", estate.ID)

	return estate, nil
}

func (s *RealEstateService) CreateRealEstate(ctx context.Context, estate *model.RealEstateInput) error {
	if estate == nil {
		return model.ErrDataIsEmpty
	}

	err := estate.Validate()
	if err != nil {
		return err
	}

	roles, err := s.userRepo.GetUserRole(ctx, estate.OwnerId)
	if err != nil {
		return err
	}

	if !model.HasRoles(roles, model.Role_Admin, model.Role_Manager) {
		return model.ErrAccessDenied
	}

	// TODO: Нужно позже выбрать тип транзакции
	tx, err := s.txRepo.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback()
	}()

	estate.Location = fmt.Sprintf("(%f, %f)", estate.Latitude, estate.Longitude)

	id, err := s.realEstateRepo.CreateRealEstate(ctx, tx, estate)
	if err != nil {
		logger.Error("s.realEstateRepo.CreateRealEstate(): ", err)
		return err
	}

	err = s.realEstateRepo.SetStatus(ctx, tx, id, estate.OwnerId, model.Status_Created)
	if err != nil {
		logger.Error("s.realEstateRepo.SetStatus(): ", err)
		return err
	}

	err = s.realEstateRepo.UpsertRealEstateURLs(ctx, tx, id, estate.NewPhotos)
	if err != nil {
		logger.Error("s.realEstateRepo.CreateRealEstateURLs(): ", err)
		return err
	}

	estate.Translations, err = s.autoTranslate(ctx, estate.Translations[0].Title, estate.Translations[0].Description, estate.Translations[0].SummaryTitle, estate.Translations[0].TermsAndConditions, estate.Translations[0].Language)
	if err != nil {
		return err
	}

	err = s.realEstateRepo.UpsertRealEstateTranslations(ctx, tx, id, estate.Translations)
	if err != nil {
		logger.Error("s.realEstateRepo.UpsertRealEstateTranslations ", err)
		return err
	}

	if len(estate.AmenityIDs) > 0 {
		for i := range estate.AmenityIDs {
			err = s.realEstateRepo.AddAmenity(ctx, tx, id, estate.AmenityIDs[i])
			if err != nil {
				logger.Error("s.realEstateRepo.AddAmenity(): ", err)
				return err
			}
		}
	}

	estate.ID = id

	err = tx.Commit()
	if err != nil {
		return err
	}
	committed = true
	return nil
}

func (s *RealEstateService) GetRealEstateTotal(ctx context.Context, filter model.RealEstateFilter, userId *int64) (int, error) {
	if userId == nil {
		filter.Status = append(filter.Status, model.Status_Available)
	}

	if err := filter.Validate(); err != nil {
		return 0, err
	}

	total, err := s.realEstateRepo.GetTotalRealEstates(ctx, filter)
	if err != nil {
		return 0, err
	}

	return total, nil
}

func (s *RealEstateService) UpdateRealEstate(ctx context.Context, updEstate *model.RealEstateInput) error {
	if updEstate == nil {
		return model.ErrDataIsEmpty
	}

	err := updEstate.Validate()
	if err != nil {
		return err
	}

	realEstate, err := s.realEstateRepo.GetRealEstate(ctx, updEstate.ID)
	if err != nil {
		return err
	}

	if realEstate == nil {
		return model.ErrNotFound
	}

	history, err := s.realEstateRepo.GetRealEstateLastHistory(ctx, realEstate.ID)
	if err != nil {
		return err
	}

	if history != nil && history.Status.ID < 0 {
		return model.ErrRealEstateStatusNotCreated
	}

	roles, err := s.userRepo.GetUserRole(ctx, updEstate.OwnerId)
	if err != nil {
		return err
	}

	if (realEstate.Manager.OwnerId != nil && updEstate.OwnerId != *realEstate.Manager.OwnerId) && !model.HasRoles(roles, model.Role_Admin) {
		return model.ErrAccessDenied
	}

	tx, err := s.txRepo.Begin(ctx)
	if err != nil {
		return err
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		_ = tx.Rollback()
	}()

	err = s.realEstateRepo.UpdateRealEstate(ctx, tx, updEstate)
	if err != nil {
		logger.Error("s.realEstateRepo.UpdateRealEstate(): ", err)
		return err
	}

	if len(updEstate.NewPhotos) > 0 {
		err = s.realEstateRepo.UpsertRealEstateURLs(ctx, tx, updEstate.ID, updEstate.NewPhotos)
		if err != nil {
			logger.Error("s.realEstateRepo.UpsertRealEstateURLs(): ", err)
			return err
		}
	}

	var deletedPhotoURLs []string
	if len(updEstate.DeletedPhotos) > 0 {
		var photos []model.RealEstatePhoto
		photos, err = s.realEstateRepo.GetRealEstateURLsByIDs(ctx, updEstate.ID, updEstate.DeletedPhotos)
		if err != nil {
			logger.Error("s.realEstateRepo.GetRealEstateURLsByIDs(): ", err)
			return err
		}

		deletedPhotoURLs = extractPhotoURLs(photos)

		err = s.realEstateRepo.DeleteSelectedRealEstatePhotos(ctx, tx, updEstate.ID, updEstate.DeletedPhotos)
		if err != nil {
			logger.Error("s.realEstateRepo.DeleteSelectedRealEstatePhotos(): ", err)
			return err
		}
	}

	estateTranslations, err := s.realEstateRepo.GetRealEstateTranslations(ctx, updEstate.ID)
	if err != nil {
		logger.Error("s.realEstateRepo.GetRealEstateTranslations(): ", err)
	}

	var oldEstateName, oldEstateDescription, oldSummaryTitle, oldTermsAndConditions string

	for _, t := range estateTranslations {
		if t.Language.Code == "ru" {
			oldEstateName = t.Name
			oldEstateDescription = t.Description
			oldSummaryTitle = t.SummaryTitle
			oldTermsAndConditions = t.TermsAndConditions
			break
		}
	}

	for _, t := range updEstate.Translations {
		if t.Language.Code == "ru" {
			if t.Title != oldEstateName || t.Description != oldEstateDescription || t.SummaryTitle != oldSummaryTitle || oldTermsAndConditions != t.TermsAndConditions {
				translations, err := s.autoTranslate(ctx, t.Title, t.Description, t.SummaryTitle, t.TermsAndConditions, t.Language)
				if err != nil {
					return err
				}

				err = s.realEstateRepo.UpsertRealEstateTranslations(ctx, tx, updEstate.ID, translations)
				if err != nil {
					logger.Error("s.realEstateRepo.UpsertRealEstateTranslations(): ", err)
					return err
				}
			}
			break
		}
	}

	if len(updEstate.AmenityIDs) > 0 {
		existingAmenities, err := s.realEstateRepo.GetAmenitiesByRealEstateID(ctx, updEstate.ID)
		if err != nil {
			return err
		}

		existingAmenityMap := make(map[int64]struct{}, len(existingAmenities))
		for _, id := range existingAmenities {
			existingAmenityMap[id] = struct{}{}
		}

		newAmenityMap := make(map[int64]struct{}, len(updEstate.AmenityIDs))
		for _, id := range updEstate.AmenityIDs {
			newAmenityMap[id] = struct{}{}
		}

		for _, existingID := range existingAmenities {
			if _, stillExists := newAmenityMap[existingID]; !stillExists {
				if err := s.realEstateRepo.RemoveAmenity(ctx, tx, updEstate.ID, existingID); err != nil {
					logger.Error("s.realEstateRepo.RemoveAmenity(): ", err)
					return err
				}
			}
		}

		for _, newID := range updEstate.AmenityIDs {
			if _, alreadyExists := existingAmenityMap[newID]; !alreadyExists {
				if err := s.realEstateRepo.AddAmenity(ctx, tx, updEstate.ID, newID); err != nil {
					logger.Error("s.realEstateRepo.AddAmenity(): ", err)
					return err
				}
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	committed = true

	if len(deletedPhotoURLs) > 0 {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := s.fileService.DeleteByURLs(bgCtx, deletedPhotoURLs); err != nil {
				logger.Error("s.fileService.DeleteByURLs(): ", err)
			}
		}()
	}

	return nil
}

func (s *RealEstateService) DeleteRealEstate(ctx context.Context, id, userId int64) error {
	if id <= 0 {
		return model.ErrIdMustBeGreaterThanZero
	}

	estate, err := s.realEstateRepo.GetRealEstate(ctx, id)
	if err != nil {
		logger.Error("s.realEstateRepo.GetRealEstate(): ", err)
		return err
	}

	roles, err := s.userRepo.GetUserRole(ctx, userId)
	if err != nil {
		logger.Error("s.userRepo.GetUserRole(): ", err)
		return err
	}

	if estate.Manager.ID != userId && !model.HasRoles(roles, model.Role_Admin) {
		return model.ErrAccessDenied
	}

	tx, err := s.txRepo.Begin(ctx)
	if err != nil {
		logger.Error("s.txRepo.Begin(): ", err)
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	photos, err := s.realEstateRepo.GetRealEstateURLs(ctx, estate.ID)
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		logger.Error("s.realEstateRepo.GetRealEstateURLs(): ", err)
		return err
	}
	deletedPhotoURLs := extractPhotoURLs(photos)

	err = s.realEstateRepo.DeleteRealEstatePhotos(ctx, tx, estate.ID)
	if err != nil {
		logger.Error("s.realEstateRepo.DeleteRealEstatePhotos(): ", err)
		return err
	}

	err = s.realEstateRepo.DeleteRealEstate(ctx, tx, estate.ID)
	if err != nil {
		logger.Error("s.realEstateRepo.DeleteRealEstate(): ", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	if len(deletedPhotoURLs) > 0 {
		go func() {
			bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			if err := s.fileService.DeleteByURLs(bgCtx, deletedPhotoURLs); err != nil {
				logger.Error("s.fileService.DeleteByURLs(): ", err)
			}
		}()
	}

	return nil
}

func extractPhotoURLs(photos []model.RealEstatePhoto) []string {
	if len(photos) == 0 {
		return []string{}
	}

	urls := make([]string, 0, len(photos))
	for _, photo := range photos {
		if photo.URL != "" {
			urls = append(urls, photo.URL)
		}
	}
	return urls
}

func (s *RealEstateService) UpdateStatus(ctx context.Context, estateId, userId int64, statusId int) error {

	switch statusId {
	case model.Status_Created, model.Status_Available, model.Status_Sold, model.Status_Archived:
	default:
		return model.ErrRealEstateStatus
	}

	estate, err := s.realEstateRepo.GetRealEstate(ctx, estateId)
	if err != nil {
		logger.Errorf("s.realEstateRepo.GetRealEstate() error: %v", err)
		return err
	}

	role, err := s.userRepo.GetUserRole(ctx, userId)
	if err != nil {
		logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		return err
	}

	if model.HasRoles(role, model.Role_Admin) {
	} else if estate.Manager.ID == userId {
		if statusId == model.Status_Available {
			return model.ErrAccessDenied
		}
	} else {
		return model.ErrAccessDenied
	}

	estate.LatestHistory, err = s.realEstateRepo.GetRealEstateLastHistory(ctx, estateId)
	if err != nil {
		logger.Errorf("s.realEstateRepo.GetRealEstateLastHistory() error: %v", err)
		return err
	}

	if statusId == estate.LatestHistory.Status.ID {
		return nil
	}

	tx, err := s.txRepo.Begin(ctx)
	if err != nil {
		logger.Errorf("s.txRepo.Begin() error: %v", err)
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = s.realEstateRepo.SetStatus(ctx, tx, estateId, userId, statusId)
	if err != nil {
		logger.Errorf("s.realEstateRepo.SetStatus() error: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Errorf("tx.Commit() error: %v", err)
		return err
	}

	return nil
}

func (s *RealEstateService) autoTranslate(ctx context.Context, title, description, summaryTitle, termsAndConditions string, lang model.Language) ([]model.RealEstateTranslationInput, error) {
	targetLangs := []string{"en", "de", "tr"}
	translations := make([]model.RealEstateTranslationInput, 0, len(targetLangs)+1)

	group, groupCtx := async.WithContext(ctx)
	sem := async.NewSemaphore(s.translateConcurrency)
	var mu sync.Mutex

	for _, langCode := range targetLangs {
		code := langCode
		group.Go(async.WithTiming("translate_"+code, func(ctx context.Context) error {
			if err := sem.Acquire(ctx); err != nil {
				return err
			}
			defer sem.Release()

			langID, ok := model.LanguageMap[code]
			if !ok {
				return fmt.Errorf("unknown language code: %s", code)
			}

			tTitle, err := s.translator.Translate(groupCtx, title, code)
			if err != nil {
				return err
			}

			tDesc, err := s.translator.Translate(groupCtx, description, code)
			if err != nil {
				return err
			}

			sTitle, err := s.translator.Translate(groupCtx, summaryTitle, code)
			if err != nil {
				return err
			}

			tAndConditions, err := s.translator.Translate(groupCtx, termsAndConditions, code)
			if err != nil {
				return err
			}

			mu.Lock()
			translations = append(translations, model.RealEstateTranslationInput{
				Title:              tTitle,
				Description:        tDesc,
				Language:           model.Language{Code: code, ID: langID},
				SummaryTitle:       sTitle,
				TermsAndConditions: tAndConditions,
			})
			mu.Unlock()

			return nil
		}))
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}

	translations = append(translations, model.RealEstateTranslationInput{
		Title:              title,
		Description:        description,
		Language:           lang,
		SummaryTitle:       summaryTitle,
		TermsAndConditions: termsAndConditions,
	})

	return translations, nil
}
