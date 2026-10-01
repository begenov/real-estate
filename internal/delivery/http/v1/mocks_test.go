package v1

import (
	"context"
	"time"

	"github.com/begenov/real-estate/internal/model"
	redisRepo "github.com/begenov/real-estate/internal/repository/redis"
	"github.com/begenov/real-estate/internal/service"
	"github.com/begenov/real-estate/pkg/auth"
)

// ── compile-time interface checks ─────────────────────────────────────────────

var _ service.IUserService = (*mockUserSvc)(nil)
var _ service.IRealEstateService = (*mockRealEstateSvc)(nil)
var _ service.ICollectionService = (*mockCollectionSvc)(nil)
var _ service.IMinioService = (*mockMinioSvc)(nil)
var _ auth.TokenManager = (*mockTokenMgr)(nil)
var _ service.IAmenityService = (*mockAmenitySvc)(nil)
var _ service.ILocationService = (*mockLocationSvc)(nil)
var _ service.IEmailService = (*mockEmailSvc)(nil)
var _ service.IPageService = (*mockPageSvc)(nil)
var _ service.IBlockService = (*mockBlockSvc)(nil)
var _ redisRepo.IRedisRepo = (*mockRedis)(nil)

// ── IUserService ──────────────────────────────────────────────────────────────

type mockUserSvc struct {
	SignInFn            func(ctx context.Context, username, password string) (*model.Token, error)
	VerifyAccessTokenFn func(ctx context.Context, userId int64, accessUUID, sessionUuid string) error
	RefreshTokenFn      func(ctx context.Context, refreshToken string) (*model.Token, error)
	LogoutFn            func(ctx context.Context, token *model.TokenDetails) error
	CreateUserFn        func(ctx context.Context, reqUser *model.UserCreateInput) error
	GetUsersFn          func(ctx context.Context, filter *model.UsersFilter, currentUserId int64) (model.Users, int, error)
	GetUserFn           func(ctx context.Context, filter *model.UserFilter, currentUserId int64) (*model.User, error)
	UpdateUserFn        func(ctx context.Context, reqUser *model.UserCreateInput) error
	DeleteUserFn        func(ctx context.Context, adminUserID, userID int64) error
}

func (m *mockUserSvc) SignIn(ctx context.Context, username, password string) (*model.Token, error) {
	return m.SignInFn(ctx, username, password)
}
func (m *mockUserSvc) VerifyAccessToken(ctx context.Context, userId int64, accessUUID, sessionUuid string) error {
	return m.VerifyAccessTokenFn(ctx, userId, accessUUID, sessionUuid)
}
func (m *mockUserSvc) RefreshToken(ctx context.Context, refreshToken string) (*model.Token, error) {
	return m.RefreshTokenFn(ctx, refreshToken)
}
func (m *mockUserSvc) Logout(ctx context.Context, token *model.TokenDetails) error {
	return m.LogoutFn(ctx, token)
}
func (m *mockUserSvc) CreateUser(ctx context.Context, reqUser *model.UserCreateInput) error {
	return m.CreateUserFn(ctx, reqUser)
}
func (m *mockUserSvc) GetUsers(ctx context.Context, filter *model.UsersFilter, currentUserId int64) (model.Users, int, error) {
	return m.GetUsersFn(ctx, filter, currentUserId)
}
func (m *mockUserSvc) GetUser(ctx context.Context, filter *model.UserFilter, currentUserId int64) (*model.User, error) {
	return m.GetUserFn(ctx, filter, currentUserId)
}
func (m *mockUserSvc) UpdateUser(ctx context.Context, reqUser *model.UserCreateInput) error {
	return m.UpdateUserFn(ctx, reqUser)
}
func (m *mockUserSvc) DeleteUser(ctx context.Context, adminUserID, userID int64) error {
	return m.DeleteUserFn(ctx, adminUserID, userID)
}

// ── IRealEstateService ────────────────────────────────────────────────────────

type mockRealEstateSvc struct {
	CreateRealEstateFn   func(ctx context.Context, estate *model.RealEstateInput) error
	GetRealEstatesFn     func(ctx context.Context, filter model.RealEstateFilter, userId *int64) (model.RealEstates, int, error)
	GetRealEstateFn      func(ctx context.Context, estateId int64, userId *int64) (*model.RealEstate, error)
	GetRealEstateTotalFn func(ctx context.Context, filter model.RealEstateFilter, userId *int64) (int, error)
	UpdateRealEstateFn   func(ctx context.Context, estate *model.RealEstateInput) error
	DeleteRealEstateFn   func(ctx context.Context, id, userId int64) error
	UpdateStatusFn       func(ctx context.Context, estateId, userId int64, statusId int) error
}

func (m *mockRealEstateSvc) CreateRealEstate(ctx context.Context, estate *model.RealEstateInput) error {
	return m.CreateRealEstateFn(ctx, estate)
}
func (m *mockRealEstateSvc) GetRealEstates(ctx context.Context, filter model.RealEstateFilter, userId *int64) (model.RealEstates, int, error) {
	return m.GetRealEstatesFn(ctx, filter, userId)
}
func (m *mockRealEstateSvc) GetRealEstate(ctx context.Context, estateId int64, userId *int64) (*model.RealEstate, error) {
	return m.GetRealEstateFn(ctx, estateId, userId)
}
func (m *mockRealEstateSvc) GetRealEstateTotal(ctx context.Context, filter model.RealEstateFilter, userId *int64) (int, error) {
	return m.GetRealEstateTotalFn(ctx, filter, userId)
}
func (m *mockRealEstateSvc) UpdateRealEstate(ctx context.Context, estate *model.RealEstateInput) error {
	return m.UpdateRealEstateFn(ctx, estate)
}
func (m *mockRealEstateSvc) DeleteRealEstate(ctx context.Context, id, userId int64) error {
	return m.DeleteRealEstateFn(ctx, id, userId)
}
func (m *mockRealEstateSvc) UpdateStatus(ctx context.Context, estateId, userId int64, statusId int) error {
	return m.UpdateStatusFn(ctx, estateId, userId, statusId)
}

// ── ICollectionService ────────────────────────────────────────────────────────

type mockCollectionSvc struct {
	GetCollectionsFn    func(ctx context.Context, filter *model.CollectionFilter, userId *int64) ([]model.Collection, int, error)
	CreateCollectionFn  func(ctx context.Context, collection *model.CollectionInput) error
	UpdateCollectionFn  func(ctx context.Context, updCollection *model.CollectionInput) error
	DeleteCollectionFn  func(ctx context.Context, collectionId, userId int64) error
	GetCollectionFn     func(ctx context.Context, collectionId int64, userId *int64, page, rows int) (*model.Collection, int, error)
	GenerateUUIDFn      func(ctx context.Context, collectionId, userId int64) (string, error)
	GetCollectionUUIDFn func(ctx context.Context, uuid string, page, rows int) (*model.Collection, int, error)
	UpdateStatusFn      func(ctx context.Context, collectionId, userId int64, statusId int) error
}

func (m *mockCollectionSvc) GetCollections(ctx context.Context, filter *model.CollectionFilter, userId *int64) ([]model.Collection, int, error) {
	return m.GetCollectionsFn(ctx, filter, userId)
}
func (m *mockCollectionSvc) CreateCollection(ctx context.Context, collection *model.CollectionInput) error {
	return m.CreateCollectionFn(ctx, collection)
}
func (m *mockCollectionSvc) UpdateCollection(ctx context.Context, updCollection *model.CollectionInput) error {
	return m.UpdateCollectionFn(ctx, updCollection)
}
func (m *mockCollectionSvc) DeleteCollection(ctx context.Context, collectionId, userId int64) error {
	return m.DeleteCollectionFn(ctx, collectionId, userId)
}
func (m *mockCollectionSvc) GetCollection(ctx context.Context, collectionId int64, userId *int64, page, rows int) (*model.Collection, int, error) {
	return m.GetCollectionFn(ctx, collectionId, userId, page, rows)
}
func (m *mockCollectionSvc) GenerateUUID(ctx context.Context, collectionId, userId int64) (string, error) {
	return m.GenerateUUIDFn(ctx, collectionId, userId)
}
func (m *mockCollectionSvc) GetCollectionUUID(ctx context.Context, uuid string, page, rows int) (*model.Collection, int, error) {
	return m.GetCollectionUUIDFn(ctx, uuid, page, rows)
}
func (m *mockCollectionSvc) UpdateStatus(ctx context.Context, collectionId, userId int64, statusId int) error {
	return m.UpdateStatusFn(ctx, collectionId, userId, statusId)
}

// ── IMinioService ─────────────────────────────────────────────────────────────

type mockMinioSvc struct {
	UploadFn          func(ctx context.Context, input *model.UploadInput) (*model.File, error)
	UploadUserPhotoFn func(ctx context.Context, input *model.UploadInput) (string, error)
	DeleteByURLsFn    func(ctx context.Context, urls []string) error
}

func (m *mockMinioSvc) Upload(ctx context.Context, input *model.UploadInput) (*model.File, error) {
	return m.UploadFn(ctx, input)
}
func (m *mockMinioSvc) UploadUserPhoto(ctx context.Context, input *model.UploadInput) (string, error) {
	return m.UploadUserPhotoFn(ctx, input)
}
func (m *mockMinioSvc) DeleteByURLs(ctx context.Context, urls []string) error {
	return m.DeleteByURLsFn(ctx, urls)
}

// ── auth.TokenManager ─────────────────────────────────────────────────────────

type mockTokenMgr struct {
	ParseAccessTokenFn  func(accessToken string) (*model.TokenDetails, error)
	ParseRefreshTokenFn func(refreshToken string) (*model.TokenDetails, error)
	NewTokenDetailsFn   func(userId int64) (*model.TokenDetails, error)
}

func (m *mockTokenMgr) ParseAccessToken(accessToken string) (*model.TokenDetails, error) {
	return m.ParseAccessTokenFn(accessToken)
}
func (m *mockTokenMgr) ParseRefreshToken(refreshToken string) (*model.TokenDetails, error) {
	return m.ParseRefreshTokenFn(refreshToken)
}
func (m *mockTokenMgr) NewTokenDetails(userId int64) (*model.TokenDetails, error) {
	return m.NewTokenDetailsFn(userId)
}

// ── IAmenityService ───────────────────────────────────────────────────────────

type mockAmenitySvc struct {
	CreateFn   func(ctx context.Context, amenity *model.Amenity) error
	UpdateFn   func(ctx context.Context, amenity *model.Amenity) error
	DeleteFn   func(ctx context.Context, amenityId int64) error
	GetAllFn   func(ctx context.Context) ([]*model.Amenity, error)
	GetByIDFn  func(ctx context.Context, id int64) (*model.Amenity, error)
	GetByIDsFn func(ctx context.Context, ids []int64) ([]*model.Amenity, error)
}

func (m *mockAmenitySvc) Create(ctx context.Context, amenity *model.Amenity) error {
	return m.CreateFn(ctx, amenity)
}
func (m *mockAmenitySvc) Update(ctx context.Context, amenity *model.Amenity) error {
	return m.UpdateFn(ctx, amenity)
}
func (m *mockAmenitySvc) Delete(ctx context.Context, amenityId int64) error {
	return m.DeleteFn(ctx, amenityId)
}
func (m *mockAmenitySvc) GetAll(ctx context.Context) ([]*model.Amenity, error) {
	return m.GetAllFn(ctx)
}
func (m *mockAmenitySvc) GetByID(ctx context.Context, id int64) (*model.Amenity, error) {
	return m.GetByIDFn(ctx, id)
}
func (m *mockAmenitySvc) GetByIDs(ctx context.Context, ids []int64) ([]*model.Amenity, error) {
	return m.GetByIDsFn(ctx, ids)
}

// ── ILocationService ──────────────────────────────────────────────────────────

type mockLocationSvc struct {
	GetCountriesFn    func(ctx context.Context) ([]*model.Country, error)
	CreateCountryFn   func(ctx context.Context, country *model.Country) error
	UpdateCountryFn   func(ctx context.Context, country *model.Country) error
	DeleteCountryFn   func(ctx context.Context, countryId int64) error
	GetCountryByIDFn  func(ctx context.Context, id int64) (*model.Country, error)
	GetRegionsFn      func(ctx context.Context, filter *model.RegionFilter) ([]*model.Region, int, error)
	CreateRegionFn    func(ctx context.Context, region *model.Region) error
	UpdateRegionFn    func(ctx context.Context, region *model.Region) error
	DeleteRegionFn    func(ctx context.Context, regionId int64) error
	GetRegionByIDFn   func(ctx context.Context, id int64) (*model.Region, error)
	GetDistrictsFn    func(ctx context.Context, filter *model.DistrictFilter) ([]*model.District, int, error)
	CreateDistrictFn  func(ctx context.Context, district *model.District) error
	UpdateDistrictFn  func(ctx context.Context, district *model.District) error
	DeleteDistrictFn  func(ctx context.Context, districtId int64) error
	GetDistrictByIDFn func(ctx context.Context, id int64) (*model.District, error)
}

func (m *mockLocationSvc) GetCountries(ctx context.Context) ([]*model.Country, error) {
	return m.GetCountriesFn(ctx)
}
func (m *mockLocationSvc) CreateCountry(ctx context.Context, country *model.Country) error {
	return m.CreateCountryFn(ctx, country)
}
func (m *mockLocationSvc) UpdateCountry(ctx context.Context, country *model.Country) error {
	return m.UpdateCountryFn(ctx, country)
}
func (m *mockLocationSvc) DeleteCountry(ctx context.Context, countryId int64) error {
	return m.DeleteCountryFn(ctx, countryId)
}
func (m *mockLocationSvc) GetCountryByID(ctx context.Context, id int64) (*model.Country, error) {
	return m.GetCountryByIDFn(ctx, id)
}
func (m *mockLocationSvc) GetRegions(ctx context.Context, filter *model.RegionFilter) ([]*model.Region, int, error) {
	return m.GetRegionsFn(ctx, filter)
}
func (m *mockLocationSvc) CreateRegion(ctx context.Context, region *model.Region) error {
	return m.CreateRegionFn(ctx, region)
}
func (m *mockLocationSvc) UpdateRegion(ctx context.Context, region *model.Region) error {
	return m.UpdateRegionFn(ctx, region)
}
func (m *mockLocationSvc) DeleteRegion(ctx context.Context, regionId int64) error {
	return m.DeleteRegionFn(ctx, regionId)
}
func (m *mockLocationSvc) GetRegionByID(ctx context.Context, id int64) (*model.Region, error) {
	return m.GetRegionByIDFn(ctx, id)
}
func (m *mockLocationSvc) GetDistricts(ctx context.Context, filter *model.DistrictFilter) ([]*model.District, int, error) {
	return m.GetDistrictsFn(ctx, filter)
}
func (m *mockLocationSvc) CreateDistrict(ctx context.Context, district *model.District) error {
	return m.CreateDistrictFn(ctx, district)
}
func (m *mockLocationSvc) UpdateDistrict(ctx context.Context, district *model.District) error {
	return m.UpdateDistrictFn(ctx, district)
}
func (m *mockLocationSvc) DeleteDistrict(ctx context.Context, districtId int64) error {
	return m.DeleteDistrictFn(ctx, districtId)
}
func (m *mockLocationSvc) GetDistrictByID(ctx context.Context, id int64) (*model.District, error) {
	return m.GetDistrictByIDFn(ctx, id)
}

// ── IEmailService ─────────────────────────────────────────────────────────────

type mockEmailSvc struct {
	SendContactFormFn func(name, phone, message string) error
	SendTourFormFn    func(name, phone string, desiredDate time.Time) error
}

func (m *mockEmailSvc) SendContactForm(name, phone, message string) error {
	return m.SendContactFormFn(name, phone, message)
}
func (m *mockEmailSvc) SendTourForm(name, phone string, desiredDate time.Time) error {
	return m.SendTourFormFn(name, phone, desiredDate)
}

// ── IPageService ──────────────────────────────────────────────────────────────

type mockPageSvc struct {
	GetPageFn  func(ctx context.Context, id int64) (*model.Page, error)
	GetPagesFn func(ctx context.Context) ([]*model.Page, error)
	CreateFn   func(ctx context.Context, page *model.Page) error
	UpdateFn   func(ctx context.Context, page *model.Page) error
	DeleteFn   func(ctx context.Context, id int64) error
}

func (m *mockPageSvc) GetPage(ctx context.Context, id int64) (*model.Page, error) {
	return m.GetPageFn(ctx, id)
}
func (m *mockPageSvc) GetPages(ctx context.Context) ([]*model.Page, error) {
	return m.GetPagesFn(ctx)
}
func (m *mockPageSvc) Create(ctx context.Context, page *model.Page) error {
	return m.CreateFn(ctx, page)
}
func (m *mockPageSvc) Update(ctx context.Context, page *model.Page) error {
	return m.UpdateFn(ctx, page)
}
func (m *mockPageSvc) Delete(ctx context.Context, id int64) error {
	return m.DeleteFn(ctx, id)
}

// ── IBlockService ─────────────────────────────────────────────────────────────

type mockBlockSvc struct {
	CreateBlockFn func(ctx context.Context, block *model.Block) error
	UpdateBlockFn func(ctx context.Context, block *model.Block) error
	GetBlockFn    func(ctx context.Context, id int64) (*model.Block, error)
	GetBlocksFn   func(ctx context.Context, pageID int64) ([]*model.Block, error)
	DeleteBlockFn func(ctx context.Context, id int64) error
}

func (m *mockBlockSvc) Create(ctx context.Context, block *model.Block) error {
	return m.CreateBlockFn(ctx, block)
}
func (m *mockBlockSvc) Update(ctx context.Context, block *model.Block) error {
	return m.UpdateBlockFn(ctx, block)
}
func (m *mockBlockSvc) GetBlock(ctx context.Context, id int64) (*model.Block, error) {
	return m.GetBlockFn(ctx, id)
}
func (m *mockBlockSvc) GetBlocks(ctx context.Context, pageID int64) ([]*model.Block, error) {
	return m.GetBlocksFn(ctx, pageID)
}
func (m *mockBlockSvc) Delete(ctx context.Context, id int64) error {
	return m.DeleteBlockFn(ctx, id)
}

// ── IRedisRepo ────────────────────────────────────────────────────────────────

type mockRedis struct {
	SaveFn   func(ctx context.Context, key string, value interface{}, expires time.Duration) error
	GetFn    func(ctx context.Context, key string) (string, error)
	DeleteFn func(ctx context.Context, key ...string) error
}

func (m *mockRedis) Save(ctx context.Context, key string, value interface{}, expires time.Duration) error {
	return m.SaveFn(ctx, key, value, expires)
}
func (m *mockRedis) Get(ctx context.Context, key string) (string, error) {
	return m.GetFn(ctx, key)
}
func (m *mockRedis) Delete(ctx context.Context, key ...string) error {
	return m.DeleteFn(ctx, key...)
}

// ── test helpers ──────────────────────────────────────────────────────────────

const testBearerToken = "Bearer valid-test-token"

func testTokenDetails() *model.TokenDetails {
	return &model.TokenDetails{
		AccessToken: "valid-test-token",
		AccessUuid:  "access-uuid",
		SessionUuid: "session-uuid",
		AtExpires:   9999999999,
		UserId:      1,
	}
}

func alwaysValidTokenMgr() *mockTokenMgr {
	return &mockTokenMgr{
		ParseAccessTokenFn: func(_ string) (*model.TokenDetails, error) {
			return testTokenDetails(), nil
		},
		ParseRefreshTokenFn: func(_ string) (*model.TokenDetails, error) {
			return testTokenDetails(), nil
		},
		NewTokenDetailsFn: func(_ int64) (*model.TokenDetails, error) {
			return testTokenDetails(), nil
		},
	}
}

func alwaysVerifyUserSvc() *mockUserSvc {
	return &mockUserSvc{
		VerifyAccessTokenFn: func(_ context.Context, _ int64, _, _ string) error { return nil },
	}
}
