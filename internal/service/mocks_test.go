package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"github.com/begenov/real-estate/internal/repository/redis"
	"github.com/begenov/real-estate/pkg/auth"
	"github.com/begenov/real-estate/pkg/hash"
)

// ── hash ─────────────────────────────────────────────────────────────────────

type mockHash struct {
	GenerateFn func(password string) (string, error)
	CompareFn  func(hashed, plain string) error
}

func (m *mockHash) GenerateFromPassword(p string) (string, error) { return m.GenerateFn(p) }
func (m *mockHash) CompareHashAndPassword(h, p string) error      { return m.CompareFn(h, p) }

var _ hash.PasswordHasher = (*mockHash)(nil)

// ── token manager ─────────────────────────────────────────────────────────────

type mockTokenManager struct {
	NewDetailsFn        func(userId int64) (*model.TokenDetails, error)
	ParseAccessTokenFn  func(token string) (*model.TokenDetails, error)
	ParseRefreshTokenFn func(token string) (*model.TokenDetails, error)
}

func (m *mockTokenManager) NewTokenDetails(userId int64) (*model.TokenDetails, error) {
	return m.NewDetailsFn(userId)
}
func (m *mockTokenManager) ParseAccessToken(t string) (*model.TokenDetails, error) {
	return m.ParseAccessTokenFn(t)
}
func (m *mockTokenManager) ParseRefreshToken(t string) (*model.TokenDetails, error) {
	return m.ParseRefreshTokenFn(t)
}

var _ auth.TokenManager = (*mockTokenManager)(nil)

// ── redis ─────────────────────────────────────────────────────────────────────

type mockRedis struct {
	SaveFn   func(ctx context.Context, key string, value interface{}, exp time.Duration) error
	GetFn    func(ctx context.Context, key string) (string, error)
	DeleteFn func(ctx context.Context, key ...string) error
}

func (m *mockRedis) Save(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	return m.SaveFn(ctx, key, value, exp)
}
func (m *mockRedis) Get(ctx context.Context, key string) (string, error) {
	return m.GetFn(ctx, key)
}
func (m *mockRedis) Delete(ctx context.Context, key ...string) error {
	return m.DeleteFn(ctx, key...)
}

var _ redis.IRedisRepo = (*mockRedis)(nil)

// ── user repo ─────────────────────────────────────────────────────────────────

type mockUserRepo struct {
	GetUserFn        func(ctx context.Context, filter *model.UserFilter) (*model.User, error)
	GetUserRoleFn    func(ctx context.Context, userId int64) ([]model.Role, error)
	UsernameExistsFn func(ctx context.Context, username string) (bool, error)
	EmailExistsFn    func(ctx context.Context, email string) (bool, error)
	CreateUserFn     func(ctx context.Context, tx *sql.Tx, user *model.User) error
	InsertUserRoleFn func(ctx context.Context, tx *sql.Tx, userID int64, roleIDs []int64) error
	GetUsersFn       func(ctx context.Context, filter *model.UsersFilter) ([]*model.User, int, error)
	UpdateUserFn     func(ctx context.Context, tx *sql.Tx, user *model.User) error
	DeleteUserFn     func(ctx context.Context, userID int64) error
}

func (m *mockUserRepo) GetUser(ctx context.Context, f *model.UserFilter) (*model.User, error) {
	return m.GetUserFn(ctx, f)
}
func (m *mockUserRepo) GetUserRole(ctx context.Context, id int64) ([]model.Role, error) {
	return m.GetUserRoleFn(ctx, id)
}
func (m *mockUserRepo) UsernameExists(ctx context.Context, u string) (bool, error) {
	return m.UsernameExistsFn(ctx, u)
}
func (m *mockUserRepo) EmailExists(ctx context.Context, e string) (bool, error) {
	return m.EmailExistsFn(ctx, e)
}
func (m *mockUserRepo) CreateUser(ctx context.Context, tx *sql.Tx, u *model.User) error {
	return m.CreateUserFn(ctx, tx, u)
}
func (m *mockUserRepo) InsertUserRole(ctx context.Context, tx *sql.Tx, id int64, roles []int64) error {
	return m.InsertUserRoleFn(ctx, tx, id, roles)
}
func (m *mockUserRepo) GetUsers(ctx context.Context, f *model.UsersFilter) ([]*model.User, int, error) {
	return m.GetUsersFn(ctx, f)
}
func (m *mockUserRepo) UpdateUser(ctx context.Context, tx *sql.Tx, u *model.User) error {
	return m.UpdateUserFn(ctx, tx, u)
}
func (m *mockUserRepo) DeleteUser(ctx context.Context, id int64) error {
	return m.DeleteUserFn(ctx, id)
}

var _ postgres.IUserRepo = (*mockUserRepo)(nil)

// ── tx repo ───────────────────────────────────────────────────────────────────

type mockTxRepo struct {
	BeginFn func(ctx context.Context) (*sql.Tx, error)
}

func (m *mockTxRepo) Begin(ctx context.Context) (*sql.Tx, error) { return m.BeginFn(ctx) }

var _ postgres.ITxRepo = (*mockTxRepo)(nil)

// ── real estate repo ──────────────────────────────────────────────────────────

type mockRealEstateRepo struct {
	GetRealEstateFn                  func(ctx context.Context, id int64) (*model.RealEstate, error)
	GetRealEstatesFn                 func(ctx context.Context, filter model.RealEstateFilter) (model.RealEstates, int, error)
	CreateRealEstateFn               func(ctx context.Context, tx *sql.Tx, re *model.RealEstateInput) (int64, error)
	UpdateRealEstateFn               func(ctx context.Context, tx *sql.Tx, re *model.RealEstateInput) error
	DeleteRealEstateFn               func(ctx context.Context, tx *sql.Tx, id int64) error
	DeleteRealEstatePhotosFn         func(ctx context.Context, tx *sql.Tx, realEstateId int64) error
	DeleteSelectedRealEstatePhotosFn func(ctx context.Context, tx *sql.Tx, realEstateId int64, photoIDs []int64) error
	GetRealEstateURLsFn              func(ctx context.Context, realEstateId int64) ([]model.RealEstatePhoto, error)
	GetRealEstateURLFn               func(ctx context.Context, realEstateId int64) (*model.RealEstatePhoto, error)
	GetRealEstateURLsByIDsFn         func(ctx context.Context, realEstateId int64, photoIDs []int64) ([]model.RealEstatePhoto, error)
	UpsertRealEstateURLsFn           func(ctx context.Context, tx *sql.Tx, realEstateId int64, photos []model.RealEstatePhoto) error
	SetStatusFn                      func(ctx context.Context, tx *sql.Tx, realEstateId, userId int64, status int) error
	GetRealEstateLastHistoryFn       func(ctx context.Context, realEstateId int64) (*model.RealEstateHistory, error)
	GetRealEstateLastHistoriesFn     func(ctx context.Context, ids []int64) (map[int64]*model.RealEstateHistory, error)
	GetTotalRealEstatesFn            func(ctx context.Context, filter model.RealEstateFilter) (int, error)
	UpsertRealEstateTranslationsFn   func(ctx context.Context, tx *sql.Tx, id int64, t []model.RealEstateTranslationInput) error
	GetRealEstateTranslationsFn      func(ctx context.Context, realEstateId int64) ([]model.RealEstateTranslation, error)
	GetRealEstateTranslationsByIDsFn func(ctx context.Context, ids []int64) (map[int64][]model.RealEstateTranslation, error)
	AddAmenityFn                     func(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error
	RemoveAmenityFn                  func(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error
	GetAmenitiesByRealEstateIDFn     func(ctx context.Context, realEstateID int64) ([]int64, error)
}

func (m *mockRealEstateRepo) GetRealEstate(ctx context.Context, id int64) (*model.RealEstate, error) {
	return m.GetRealEstateFn(ctx, id)
}
func (m *mockRealEstateRepo) GetRealEstates(ctx context.Context, f model.RealEstateFilter) (model.RealEstates, int, error) {
	return m.GetRealEstatesFn(ctx, f)
}
func (m *mockRealEstateRepo) CreateRealEstate(ctx context.Context, tx *sql.Tx, re *model.RealEstateInput) (int64, error) {
	return m.CreateRealEstateFn(ctx, tx, re)
}
func (m *mockRealEstateRepo) UpdateRealEstate(ctx context.Context, tx *sql.Tx, re *model.RealEstateInput) error {
	return m.UpdateRealEstateFn(ctx, tx, re)
}
func (m *mockRealEstateRepo) DeleteRealEstate(ctx context.Context, tx *sql.Tx, id int64) error {
	return m.DeleteRealEstateFn(ctx, tx, id)
}
func (m *mockRealEstateRepo) DeleteRealEstatePhotos(ctx context.Context, tx *sql.Tx, realEstateId int64) error {
	return m.DeleteRealEstatePhotosFn(ctx, tx, realEstateId)
}
func (m *mockRealEstateRepo) DeleteSelectedRealEstatePhotos(ctx context.Context, tx *sql.Tx, realEstateId int64, photoIDs []int64) error {
	return m.DeleteSelectedRealEstatePhotosFn(ctx, tx, realEstateId, photoIDs)
}
func (m *mockRealEstateRepo) GetRealEstateURLs(ctx context.Context, id int64) ([]model.RealEstatePhoto, error) {
	return m.GetRealEstateURLsFn(ctx, id)
}
func (m *mockRealEstateRepo) GetRealEstateURL(ctx context.Context, id int64) (*model.RealEstatePhoto, error) {
	return m.GetRealEstateURLFn(ctx, id)
}
func (m *mockRealEstateRepo) GetRealEstateURLsByIDs(ctx context.Context, realEstateId int64, photoIDs []int64) ([]model.RealEstatePhoto, error) {
	return m.GetRealEstateURLsByIDsFn(ctx, realEstateId, photoIDs)
}
func (m *mockRealEstateRepo) UpsertRealEstateURLs(ctx context.Context, tx *sql.Tx, id int64, photos []model.RealEstatePhoto) error {
	return m.UpsertRealEstateURLsFn(ctx, tx, id, photos)
}
func (m *mockRealEstateRepo) SetStatus(ctx context.Context, tx *sql.Tx, realEstateId, userId int64, status int) error {
	return m.SetStatusFn(ctx, tx, realEstateId, userId, status)
}
func (m *mockRealEstateRepo) GetRealEstateLastHistory(ctx context.Context, id int64) (*model.RealEstateHistory, error) {
	return m.GetRealEstateLastHistoryFn(ctx, id)
}
func (m *mockRealEstateRepo) GetRealEstateLastHistories(ctx context.Context, ids []int64) (map[int64]*model.RealEstateHistory, error) {
	return m.GetRealEstateLastHistoriesFn(ctx, ids)
}
func (m *mockRealEstateRepo) GetTotalRealEstates(ctx context.Context, f model.RealEstateFilter) (int, error) {
	return m.GetTotalRealEstatesFn(ctx, f)
}
func (m *mockRealEstateRepo) UpsertRealEstateTranslations(ctx context.Context, tx *sql.Tx, id int64, t []model.RealEstateTranslationInput) error {
	return m.UpsertRealEstateTranslationsFn(ctx, tx, id, t)
}
func (m *mockRealEstateRepo) GetRealEstateTranslations(ctx context.Context, id int64) ([]model.RealEstateTranslation, error) {
	return m.GetRealEstateTranslationsFn(ctx, id)
}
func (m *mockRealEstateRepo) GetRealEstateTranslationsByIDs(ctx context.Context, ids []int64) (map[int64][]model.RealEstateTranslation, error) {
	return m.GetRealEstateTranslationsByIDsFn(ctx, ids)
}
func (m *mockRealEstateRepo) AddAmenity(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error {
	return m.AddAmenityFn(ctx, tx, realEstateID, amenityID)
}
func (m *mockRealEstateRepo) RemoveAmenity(ctx context.Context, tx *sql.Tx, realEstateID, amenityID int64) error {
	return m.RemoveAmenityFn(ctx, tx, realEstateID, amenityID)
}
func (m *mockRealEstateRepo) GetAmenitiesByRealEstateID(ctx context.Context, id int64) ([]int64, error) {
	return m.GetAmenitiesByRealEstateIDFn(ctx, id)
}

var _ postgres.IRealEstateRepo = (*mockRealEstateRepo)(nil)

// ── amenity service ───────────────────────────────────────────────────────────

type mockAmenityService struct {
	GetByIDsFn func(ctx context.Context, ids []int64) ([]*model.Amenity, error)
}

func (m *mockAmenityService) Create(ctx context.Context, a *model.Amenity) error   { return nil }
func (m *mockAmenityService) Update(ctx context.Context, a *model.Amenity) error   { return nil }
func (m *mockAmenityService) Delete(ctx context.Context, id int64) error           { return nil }
func (m *mockAmenityService) GetAll(ctx context.Context) ([]*model.Amenity, error) { return nil, nil }
func (m *mockAmenityService) GetByID(ctx context.Context, id int64) (*model.Amenity, error) {
	return nil, nil
}
func (m *mockAmenityService) GetByIDs(ctx context.Context, ids []int64) ([]*model.Amenity, error) {
	return m.GetByIDsFn(ctx, ids)
}

var _ IAmenityService = (*mockAmenityService)(nil)

// ── location service ──────────────────────────────────────────────────────────

type mockLocationService struct {
	GetRegionByIDFn   func(ctx context.Context, id int64) (*model.Region, error)
	GetDistrictByIDFn func(ctx context.Context, id int64) (*model.District, error)
}

func (m *mockLocationService) GetCountries(ctx context.Context) ([]*model.Country, error) {
	return nil, nil
}
func (m *mockLocationService) CreateCountry(ctx context.Context, c *model.Country) error { return nil }
func (m *mockLocationService) UpdateCountry(ctx context.Context, c *model.Country) error { return nil }
func (m *mockLocationService) DeleteCountry(ctx context.Context, id int64) error         { return nil }
func (m *mockLocationService) GetCountryByID(ctx context.Context, id int64) (*model.Country, error) {
	return nil, nil
}
func (m *mockLocationService) GetRegions(ctx context.Context, f *model.RegionFilter) ([]*model.Region, int, error) {
	return nil, 0, nil
}
func (m *mockLocationService) CreateRegion(ctx context.Context, r *model.Region) error { return nil }
func (m *mockLocationService) UpdateRegion(ctx context.Context, r *model.Region) error { return nil }
func (m *mockLocationService) DeleteRegion(ctx context.Context, id int64) error        { return nil }
func (m *mockLocationService) GetRegionByID(ctx context.Context, id int64) (*model.Region, error) {
	return m.GetRegionByIDFn(ctx, id)
}
func (m *mockLocationService) GetDistricts(ctx context.Context, f *model.DistrictFilter) ([]*model.District, int, error) {
	return nil, 0, nil
}
func (m *mockLocationService) CreateDistrict(ctx context.Context, d *model.District) error {
	return nil
}
func (m *mockLocationService) UpdateDistrict(ctx context.Context, d *model.District) error {
	return nil
}
func (m *mockLocationService) DeleteDistrict(ctx context.Context, id int64) error { return nil }
func (m *mockLocationService) GetDistrictByID(ctx context.Context, id int64) (*model.District, error) {
	return m.GetDistrictByIDFn(ctx, id)
}

var _ ILocationService = (*mockLocationService)(nil)

// ── exchange rate service ─────────────────────────────────────────────────────

type mockExchangeRate struct {
	GetRatesFn func(ctx context.Context) (map[string]float64, error)
}

func (m *mockExchangeRate) StartUpdater(ctx context.Context) {}
func (m *mockExchangeRate) GetRates(ctx context.Context) (map[string]float64, error) {
	return m.GetRatesFn(ctx)
}

var _ IExchangeRateService = (*mockExchangeRate)(nil)

// ── translate service ─────────────────────────────────────────────────────────

type mockTranslate struct{}

func (m *mockTranslate) Translate(ctx context.Context, text, lang string) (string, error) {
	return text, nil
}

var _ ITranslateService = (*mockTranslate)(nil)

// ── minio service ─────────────────────────────────────────────────────────────

type mockMinioService struct {
	DeleteByURLsFn func(ctx context.Context, urls []string) error
}

func (m *mockMinioService) Upload(ctx context.Context, input *model.UploadInput) (*model.File, error) {
	return nil, nil
}
func (m *mockMinioService) UploadUserPhoto(ctx context.Context, input *model.UploadInput) (string, error) {
	return "", nil
}
func (m *mockMinioService) DeleteByURLs(ctx context.Context, urls []string) error {
	if m.DeleteByURLsFn != nil {
		return m.DeleteByURLsFn(ctx, urls)
	}
	return nil
}

var _ IMinioService = (*mockMinioService)(nil)
