package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"unicode/utf8"

	"github.com/google/uuid"
)

type ICollectionService interface {
	GetCollections(ctx context.Context, filter *model.CollectionFilter, userId *int64) ([]model.Collection, int, error)
	CreateCollection(ctx context.Context, collection *model.CollectionInput) error
	UpdateCollection(ctx context.Context, updCollection *model.CollectionInput) error
	DeleteCollection(ctx context.Context, collectionId, userId int64) error
	GetCollection(ctx context.Context, collectionId int64, userId *int64, page, rows int) (*model.Collection, int, error)
	GenerateUUID(ctx context.Context, collectionId, userId int64) (string, error)
	GetCollectionUUID(ctx context.Context, uuid string, page, rows int) (*model.Collection, int, error)
	UpdateStatus(ctx context.Context, collectionId, userId int64, statusId int) error
}

type CollectionService struct {
	collectionRepo      postgres.ICollectionRepo
	realEstateRepo      postgres.IRealEstateRepo
	userRepo            postgres.IUserRepo
	txRepo              postgres.ITxRepo
	exchangeRateService IExchangeRateService
	translateService    ITranslateService
}

func NewCollectionService(collectionRepo postgres.ICollectionRepo, realEstateRepo postgres.IRealEstateRepo,
	userRepo postgres.IUserRepo, txRepo postgres.ITxRepo, exchangeRateService IExchangeRateService,
	translateService ITranslateService) ICollectionService {
	return &CollectionService{
		collectionRepo:      collectionRepo,
		realEstateRepo:      realEstateRepo,
		userRepo:            userRepo,
		txRepo:              txRepo,
		exchangeRateService: exchangeRateService,
		translateService:    translateService,
	}
}

func (s *CollectionService) CreateCollection(ctx context.Context, collection *model.CollectionInput) error {

	if collection == nil {
		return model.ErrDataIsEmpty
	}

	err := collection.Validate()
	if err != nil {
		return err
	}

	roles, err := s.userRepo.GetUserRole(ctx, collection.OwnerId)
	if err != nil {
		logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		return err
	}

	if !model.HasRoles(roles, model.Role_Admin, model.Role_Manager) {
		return model.ErrAccessDenied
	}

	// TODO: Нужно позже выбрать тип транзакции
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

	id, err := s.collectionRepo.CreateCollection(ctx, tx, collection)
	if err != nil {
		logger.Errorf("s.collectionRepo.CreateCollection() error: %v", err)
		return err
	}

	err = s.collectionRepo.SetStatus(ctx, tx, id, collection.OwnerId, model.Status_Created)
	if err != nil {
		logger.Errorf("s.collectionRepo.SetStatus() error: %v", err)
		return err
	}

	for i := range collection.Items {

		estate, err := s.realEstateRepo.GetRealEstate(ctx, collection.Items[i])
		if err != nil || estate == nil {
			logger.Error("s.realEstateRepo.GetRealEstate() error: ", err)
			return model.ErrRealEstateCollection
		}

		estate.LatestHistory, err = s.realEstateRepo.GetRealEstateLastHistory(ctx, estate.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Error("s.realEstateRepo.GetRealEstateLastHistory(): ", err)
			return model.ErrRealEstateHistory
		}

		//if estate.LatestHistory.Status.ID != model.Status_Available {
		//	return model.ErrRealEstateStatusNotAvailable
		//}

		err = s.collectionRepo.UpsertCollectionItem(ctx, tx, id, collection.Items[i])
		if err != nil {
			logger.Errorf("s.collectionRepo.UpsertCollectionItem() error: %v", err)
			return err
		}
	}

	collection.ID = id

	for code, langID := range model.LanguageMap {
		if code == "ru" {
			err = s.collectionRepo.InsertTranslation(ctx, tx, collection.ID, langID, collection.Name, collection.Description)
			if err != nil {
				return err
			}

			continue
		}

		tTitle, err := s.translateService.Translate(ctx, collection.Name, code)
		if err != nil {
			return err
		}

		tDesc, err := s.translateService.Translate(ctx, collection.Description, code)
		if err != nil {
			return err
		}

		err = s.collectionRepo.InsertTranslation(ctx, tx, collection.ID, langID, tTitle, tDesc)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *CollectionService) UpdateCollection(ctx context.Context, updCollection *model.CollectionInput) error {
	if updCollection == nil {
		return model.ErrDataIsEmpty
	}

	err := updCollection.Validate()
	if err != nil {
		logger.Errorf("updCollection.Validate() error: %v", err)
		return err
	}

	collection, err := s.collectionRepo.GetCollection(ctx, &model.CollectionFilter{Id: &updCollection.ID})
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollection() error: %v", err)
		return err
	}

	roles, err := s.userRepo.GetUserRole(ctx, updCollection.OwnerId)
	if err != nil {
		logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		return err
	}

	if collection.Manager.ID != updCollection.OwnerId && !model.HasRoles(roles, model.Role_Admin) {
		return model.ErrAccessDenied
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

	err = s.collectionRepo.UpdateCollection(ctx, updCollection)
	if err != nil {
		logger.Errorf("s.collectionRepo.UpdateCollection() error: %v", err)
		return err
	}

	if len(updCollection.DeletedItems) > 0 {
		for i := range updCollection.DeletedItems {
			err := s.collectionRepo.DeleteCollectionItem(ctx, tx, updCollection.ID, updCollection.DeletedItems[i])
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				logger.Errorf("s.collectionRepo.DeleteCollectionItem() error: %v", err)
				return err
			}
		}
	}

	for i := range updCollection.Items {

		estate, err := s.realEstateRepo.GetRealEstate(ctx, updCollection.Items[i])
		if err != nil || estate == nil {
			logger.Error("s.realEstateRepo.GetRealEstate() error: ", err)
			return model.ErrRealEstateCollection
		}

		estate.LatestHistory, err = s.realEstateRepo.GetRealEstateLastHistory(ctx, estate.ID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			logger.Error("s.realEstateRepo.GetRealEstateLastHistory(): ", err)
			return model.ErrRealEstateHistory
		}

		//if estate.LatestHistory.Status.ID != model.Status_Available {
		//	return model.ErrRealEstateStatusNotAvailable
		//}

		err = s.collectionRepo.UpsertCollectionItem(ctx, tx, collection.ID, updCollection.Items[i])
		if err != nil {
			logger.Errorf("s.collectionRepo.UpsertCollectionItem() error: %v", err)
			return err
		}
	}

	if collection.Description != updCollection.Description || collection.Name != updCollection.Name {
		for code, langID := range model.LanguageMap {
			if code == "ru" {
				err = s.collectionRepo.InsertTranslation(ctx, tx, updCollection.ID, langID, updCollection.Name, updCollection.Description)
				if err != nil {
					return err
				}

				continue
			}

			tTitle, err := s.translateService.Translate(ctx, updCollection.Name, code)
			if err != nil {
				return err
			}

			tDesc, err := s.translateService.Translate(ctx, updCollection.Description, code)
			if err != nil {
				return err
			}

			err = s.collectionRepo.InsertTranslation(ctx, tx, updCollection.ID, langID, tTitle, tDesc)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *CollectionService) GetCollections(ctx context.Context, filter *model.CollectionFilter, userId *int64) ([]model.Collection, int, error) {
	if filter == nil {
		return nil, 0, model.ErrDataIsEmpty
	}

	err := filter.Validate()
	if err != nil {
		logger.Errorf("filter.Validate() error: %v", err)
		return nil, 0, err
	}

	if userId == nil {
		filter.Type = append(filter.Type, model.Status_Available)
	} else {
		role, err := s.userRepo.GetUserRole(ctx, *userId)
		if err != nil {
			logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
			return nil, 0, err
		}

		if !model.HasRoles(role, model.Role_Admin) {
			filter.OwnerId = userId

			filter.Type = append(filter.Type, model.Status_Available)
		}

	}

	collections, total, err := s.collectionRepo.GetCollections(ctx, filter)
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollections() error: %v", err)
		return nil, 0, err
	}

	var isAdmin = true

	//TODO необходимо выяснить у бизнеса, нужна ли недвижимость, когда мы получим коллекцию
	for i := range collections {
		//estates, _, err := s.realEstateRepo.GetRealEstates(ctx, model.RealEstateFilter{
		//	CollectionId: &collections[i].ID,
		//})
		//if err != nil {
		//	return nil, 0, err
		//}
		//
		//collections[i].RealEstates = estates

		translations, err := s.collectionRepo.GetCollectionTranslations(ctx, collections[i].ID)
		if err != nil {
			logger.Errorf("s.collectionRepo.GetCollectionTranslations() error: %v", err)
		}

		collections[i].CollectionTranslation = translations

		if userId != nil {
			history, err := s.collectionRepo.GetCollectionLastHistory(ctx, collections[i].ID)
			if err != nil {
				return nil, 0, err
			}

			collections[i].LatestHistory = history
		}

		user, err := s.userRepo.GetUser(ctx, &model.UserFilter{Id: &collections[i].Manager.ID})
		if err != nil {
			return nil, 0, err
		}

		if user == nil {
			continue
		}

		role, err := s.userRepo.GetUserRole(ctx, user.ID)
		if err != nil {
			logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		}

		if model.HasRoles(role, model.Role_Admin) {
			collections[i].Manager.IsAdmin = &isAdmin
		}

		collections[i].Manager = *user
	}

	return collections, total, nil
}

func (s *CollectionService) DeleteCollection(ctx context.Context, collectionId, userId int64) error {

	collection, err := s.collectionRepo.GetCollection(ctx, &model.CollectionFilter{Id: &collectionId})
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollection() error: %v", err)
		return err
	}

	//history, err := s.collectionRepo.GetCollectionLastHistory(ctx, collectionId)
	//if err != nil && !errors.Is(err, model.ErrNotFound) {
	//	logger.Errorf("s.collectionRepo.GetCollectionLastHistory() error: %v", err)
	//	return err
	//}

	//if history != nil && !(history.Status.ID == model.Status_Created || history.Status.ID == model.Status_Archived) {
	//	return model.ErrInvalidCollectionStatus
	//}

	roles, err := s.userRepo.GetUserRole(ctx, userId)
	if err != nil {
		logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		return err
	}

	if collection.Manager.ID != userId && !model.HasRoles(roles, model.Role_Admin) {
		return model.ErrAccessDenied
	}

	tx, err := s.txRepo.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	err = s.collectionRepo.DeleteCollectionItems(ctx, tx, collectionId)
	if err != nil {
		logger.Errorf("s.collectionRepo.DeleteCollectionItems() error: %v", err)
		return err
	}

	err = s.collectionRepo.DeleteCollection(ctx, tx, collectionId)
	if err != nil {
		logger.Errorf("s.collectionRepo.DeleteCollection() error: %v", err)
		return err
	}

	return tx.Commit()
}

func (s *CollectionService) GetCollection(ctx context.Context, collectionId int64, userId *int64, page, rows int) (*model.Collection, int, error) {
	var filter = model.CollectionFilter{
		Id: &collectionId,
	}

	if userId == nil {
		filter.Type = append(filter.Type, model.Status_Available)
	} else {
		role, err := s.userRepo.GetUserRole(ctx, *userId)
		if err != nil {
			logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
			return nil, 0, err
		}

		if !model.HasRoles(role, model.Role_Admin) {
			filter.OwnerId = userId
		}
	}

	collection, err := s.collectionRepo.GetCollection(ctx, &filter)
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollection() error: %v", err)
		return nil, 0, err
	}

	user, err := s.userRepo.GetUser(ctx, &model.UserFilter{Id: &collection.Manager.ID})
	if err != nil {
		logger.Errorf("s.userRepo.GetUser() error: %v", err)
		return nil, 0, err
	}
	if user != nil {
		var isAdmin = true

		collection.Manager = *user

		role, err := s.userRepo.GetUserRole(ctx, user.ID)
		if err != nil {
			logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		}

		if model.HasRoles(role, model.Role_Admin) {
			collection.Manager.IsAdmin = &isAdmin
		}
	}

	rates, err := s.exchangeRateService.GetRates(ctx)
	if err != nil {
		logger.Error("GetRates(): ", err)
		rates = map[string]float64{"USD": 0, "EUR": 0} // fallback
	}

	estates, total, err := s.realEstateRepo.GetRealEstates(ctx, model.RealEstateFilter{
		CollectionId: &collectionId,
		Page:         page,
		Rows:         rows,
	})
	if err != nil {
		logger.Errorf("s.realEstateRepo.GetRealEstates() error: %v", err)
		return nil, 0, err
	}

	for i := range estates {
		estates[i].Manager = collection.Manager

		urls, err := s.realEstateRepo.GetRealEstateURLs(ctx, estates[i].ID)
		if err != nil {
			logger.Errorf("s.realEstateRepo.GetRealEstateURLs() error: %v", err)
			return nil, 0, err
		}
		estates[i].Photos = urls

		estates[i].PriceUSD = estates[i].Price * rates["USD"]
		estates[i].PriceEUR = estates[i].Price * rates["EUR"]

		estates[i].SerialNumber = fmt.Sprintf("%09d", estates[i].ID)

		translations, err := s.realEstateRepo.GetRealEstateTranslations(ctx, estates[i].ID)
		if err != nil {
			logger.Errorf("s.realEstateRepo.GetRealEstateTranslations() error: %v", err)
		}

		estates[i].Translations = translations
	}

	collection.RealEstates = estates

	if userId != nil {
		history, err := s.collectionRepo.GetCollectionLastHistory(ctx, collectionId)
		if err != nil {
			return nil, 0, err
		}
		collection.LatestHistory = history
	}

	translations, err := s.collectionRepo.GetCollectionTranslations(ctx, collectionId)
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollectionTranslations() error: %v", err)
	}

	collection.CollectionTranslation = translations

	return collection, total, nil
}

func (s *CollectionService) GetCollectionUUID(ctx context.Context, uuid string, page, rows int) (*model.Collection, int, error) {

	collection, err := s.collectionRepo.GetCollectionByUUID(ctx, uuid)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.enrichCollection(ctx, collection, page, rows)
	if err != nil {
		return nil, 0, err
	}

	return collection, total, nil
}

func (s *CollectionService) GenerateUUID(ctx context.Context, collectionId, userId int64) (string, error) {

	collection, err := s.collectionRepo.GetCollection(ctx, &model.CollectionFilter{Id: &collectionId})
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollection() error: %v", err)
		return "", err
	}

	roles, err := s.userRepo.GetUserRole(ctx, userId)
	if err != nil {
		logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		return "", err
	}

	if collection.Manager.ID != userId && !model.HasRoles(roles, model.Role_Admin) {
		return "", model.ErrAccessDenied
	}

	if collection.UUID != nil && utf8.RuneCountInString(*collection.UUID) > 0 {
		return *collection.UUID, nil
	}

	newUUID := uuid.New().String()

	err = s.collectionRepo.SetCollectionUUID(ctx, collectionId, newUUID)
	if err != nil {
		logger.Errorf("s.collectionRepo.SetCollectionUUID() error: %v", err)
		return "", err
	}

	return newUUID, nil
}

func (s *CollectionService) enrichCollection(ctx context.Context, collection *model.Collection, page, rows int) (int, error) {
	user, err := s.userRepo.GetUser(ctx, &model.UserFilter{Id: &collection.Manager.ID})
	if err != nil {
		return 0, err
	}
	if user != nil {

		var isAdmin = true

		collection.Manager = *user

		role, err := s.userRepo.GetUserRole(ctx, user.ID)
		if err != nil {
			logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		}

		if model.HasRoles(role, model.Role_Admin) {
			collection.Manager.IsAdmin = &isAdmin
		}

	}

	rates, err := s.exchangeRateService.GetRates(ctx)
	if err != nil {
		logger.Error("GetRates(): ", err)
		rates = map[string]float64{"USD": 0, "EUR": 0} // fallback
	}

	estates, total, err := s.realEstateRepo.GetRealEstates(ctx, model.RealEstateFilter{
		CollectionId: &collection.ID,
		Page:         page,
		Rows:         rows,
	})
	if err != nil {
		return 0, err
	}

	for i := range estates {
		estates[i].Manager = collection.Manager

		urls, err2 := s.realEstateRepo.GetRealEstateURLs(ctx, estates[i].ID)
		if err2 != nil {
			return 0, err2
		}
		estates[i].Photos = urls

		estates[i].PriceUSD = estates[i].Price * rates["USD"]
		estates[i].PriceEUR = estates[i].Price * rates["EUR"]

		estates[i].SerialNumber = fmt.Sprintf("%09d", estates[i].ID)

		translations, err := s.realEstateRepo.GetRealEstateTranslations(ctx, estates[i].ID)
		if err != nil {
			logger.Errorf("s.realEstateRepo.GetRealEstateTranslations() error: %v", err)
		}

		estates[i].Translations = translations
	}

	collection.RealEstates = estates

	translations, err := s.collectionRepo.GetCollectionTranslations(ctx, collection.ID)
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollectionTranslations() error: %v", err)
	}

	collection.CollectionTranslation = translations

	return total, nil
}

func (s *CollectionService) UpdateStatus(ctx context.Context, collectionId, userId int64, statusId int) error {

	switch statusId {
	case model.Status_Created, model.Status_Available, model.Status_Archived:
	default:
		return model.ErrInvalidCollectionStatus
	}

	var filter = model.CollectionFilter{
		Id: &collectionId,
	}

	collection, err := s.collectionRepo.GetCollection(ctx, &filter)
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollection() error: %v", err)
		return err
	}

	role, err := s.userRepo.GetUserRole(ctx, userId)
	if err != nil {
		logger.Errorf("s.userRepo.GetUserRole() error: %v", err)
		return err
	}

	if model.HasRoles(role, model.Role_Admin) {
	} else if collection.Manager.ID == userId {
		if statusId == model.Status_Available {
			return model.ErrAccessDenied
		}
	} else {
		return model.ErrAccessDenied
	}

	collection.LatestHistory, err = s.collectionRepo.GetCollectionLastHistory(ctx, collectionId)
	if err != nil {
		logger.Errorf("s.collectionRepo.GetCollectionLastHistory() error: %v", err)
		return err
	}

	//if statusId < collection.LatestHistory.Status.ID {
	//	return model.ErrInvalidCollectionStatus
	//}

	//if statusId == collection.LatestHistory.Status.ID {
	//	return model.ErrInvalidCollectionStatus
	//}

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

	err = s.collectionRepo.SetStatus(ctx, tx, collectionId, userId, statusId)
	if err != nil {
		logger.Errorf("s.collectionRepo.SetStatus() error: %v", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Errorf("tx.Commit() error: %v", err)
		return err
	}

	return nil
}
