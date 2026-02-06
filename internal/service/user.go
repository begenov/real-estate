package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/begenov/real-estate/internal/logger"
	"github.com/begenov/real-estate/internal/model"
	"github.com/begenov/real-estate/internal/repository/postgres"
	"github.com/begenov/real-estate/internal/repository/redis"
	"github.com/begenov/real-estate/pkg/auth"
	"github.com/begenov/real-estate/pkg/hash"
	"strings"
	"time"
	"unicode/utf8"
)

type IUserService interface {
	SignIn(ctx context.Context, username string, password string) (*model.Token, error)
	VerifyAccessToken(ctx context.Context, userId int64, accessUUID, sessionUuid string) error
	RefreshToken(ctx context.Context, refreshToken string) (*model.Token, error)
	Logout(ctx context.Context, token *model.TokenDetails) error
	CreateUser(ctx context.Context, reqUser *model.UserCreateInput) error

	GetUsers(ctx context.Context, filter *model.UsersFilter, currentUserId int64) (model.Users, int, error)
	GetUser(ctx context.Context, filter *model.UserFilter, currentUserId int64) (*model.User, error)
	UpdateUser(ctx context.Context, reqUser *model.UserCreateInput) error
	DeleteUser(ctx context.Context, adminUserID, userID int64) error
}

type UserService struct {
	hash     hash.PasswordHasher
	token    auth.TokenManager
	userRepo postgres.IUserRepo
	redis    redis.IRedisRepo
	txRepo   postgres.ITxRepo
}

func NewUserService(hash hash.PasswordHasher, token auth.TokenManager, userRepo postgres.IUserRepo, redis redis.IRedisRepo, txRepo postgres.ITxRepo) IUserService {
	return &UserService{
		hash:     hash,
		token:    token,
		userRepo: userRepo,
		redis:    redis,
		txRepo:   txRepo,
	}
}

func (s *UserService) SignIn(ctx context.Context, username string, password string) (*model.Token, error) {

	if utf8.RuneCountInString(strings.TrimSpace(password)) < 8 {
		return nil, model.ErrPasswordTooShort
	}

	var filter = model.UserFilter{
		Username: &username,
	}

	user, err := s.userRepo.GetUser(ctx, &filter)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, model.ErrSignIn
	} else if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, model.ErrUserNotActive
	}

	if user.Password == nil {
		return nil, model.ErrSignIn
	}

	err = s.hash.CompareHashAndPassword(*user.Password, password)
	if err != nil {
		return nil, model.ErrSignIn
	}

	return s.createSession(ctx, user.ID)
}

func (s *UserService) CreateUser(ctx context.Context, reqUser *model.UserCreateInput) error {

	if reqUser.Password == nil || utf8.RuneCountInString(strings.TrimSpace(*reqUser.Password)) < 8 {
		return model.ErrPasswordTooShort
	}

	passwordHash, err := s.hash.GenerateFromPassword(*reqUser.Password)
	if err != nil {
		logger.Error("GenerateFromPassword(): ", err)
		return err
	}

	err = reqUser.Validate()
	if err != nil {
		logger.Error("reqUser.Validate(): ", err)
		return err
	}

	if err := s.checkUsernameAndEmailExists(ctx, reqUser.Username, reqUser.Email); err != nil {
		logger.Error("checkUsernameAndEmailExists(): ", err)
		return err
	}

	var user = model.User{
		Username:   &reqUser.Username,
		Email:      &reqUser.Email,
		FirstName:  reqUser.FirstName,
		LastName:   reqUser.LastName,
		MiddleName: reqUser.MiddleName,
		Phone:      reqUser.Phone,
		Password:   &passwordHash,
		OwnerId:    reqUser.OwnerId,
		PhotoURL:   reqUser.PhotoURL,
	}

	tx, err := s.txRepo.Begin(ctx)
	if err != nil {
		logger.Error("Begin(): ", err)
		return err
	}

	err = s.userRepo.CreateUser(ctx, tx, &user)
	if err != nil {
		logger.Error("CreateUser(): ", err)
		return err
	}

	err = s.userRepo.InsertUserRole(ctx, tx, user.ID, []int64{model.Role_Manager_ID})
	if err != nil {
		logger.Error("InsertUserRole(): ", err)
		return err
	}

	reqUser.ID = user.ID

	return tx.Commit()
}

func (s *UserService) UpdateUser(ctx context.Context, reqUser *model.UserCreateInput) error {

	user, err := s.userRepo.GetUser(ctx, &model.UserFilter{Id: &reqUser.ID})
	if err != nil || user == nil {
		logger.Error("GetUser(): ", err)
		return err
	}

	if reqUser.Password != nil && len(*reqUser.Password) > 0 {
		if utf8.RuneCountInString(strings.TrimSpace(*reqUser.Password)) < 8 {
			return model.ErrPasswordTooShort
		}

		passwordHash, err := s.hash.GenerateFromPassword(*reqUser.Password)
		if err != nil {
			logger.Error("GenerateFromPassword(): ", err)
			return err
		}

		reqUser.Password = &passwordHash
	} else {
		reqUser.Password = nil
	}

	err = reqUser.Validate()
	if err != nil {
		logger.Error("reqUser.Validate(): ", err)
		return err
	}

	needToCheckUsername := false
	needToCheckEmail := false
	if reqUser.Username != "" && (user.Username == nil || reqUser.Username != *user.Username) {
		needToCheckUsername = true
	}
	if reqUser.Email != "" && (user.Email == nil || reqUser.Email != *user.Email) {
		needToCheckEmail = true
	}

	if needToCheckUsername {
		if exists, _ := s.userRepo.UsernameExists(ctx, reqUser.Username); exists {
			return model.ErrUsernameAlreadyExists
		}
	}

	if needToCheckEmail {
		if exists, _ := s.userRepo.EmailExists(ctx, reqUser.Email); exists {
			return model.ErrEmailAlreadyExists
		}
	}

	tx, err := s.txRepo.Begin(ctx)
	if err != nil {
		logger.Error("Begin(): ", err)
		return err
	}

	committed := false
	defer func() {
		if committed {
			return
		}
		if err := tx.Rollback(); err != nil {
			logger.Error("tx.Rollback(): ", err)
		}
	}()

	var updUser = model.User{
		ID:         reqUser.ID,
		Username:   &reqUser.Username,
		Email:      &reqUser.Email,
		FirstName:  reqUser.FirstName,
		LastName:   reqUser.LastName,
		MiddleName: reqUser.MiddleName,
		Phone:      reqUser.Phone,
		Password:   reqUser.Password,
		OwnerId:    reqUser.OwnerId,
		PhotoURL:   reqUser.PhotoURL,
		IsActive:   reqUser.IsActive,
	}

	err = s.userRepo.UpdateUser(ctx, tx, &updUser)
	if err != nil {
		logger.Error("UpdateUser(): ", err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("tx.Commit(): ", err)
		return err
	}
	committed = true

	return nil
}

func (s *UserService) GetUsers(ctx context.Context, filter *model.UsersFilter, currentUserId int64) (model.Users, int, error) {

	role, err := s.userRepo.GetUserRole(ctx, currentUserId)
	if err != nil {
		return nil, 0, err
	}

	if !model.HasRoles(role, model.Role_Admin) {
		filter.UserId = &currentUserId
	} else {
		var roleId = model.Role_Manager_ID
		filter.RoleId = &roleId
	}

	users, total, err := s.userRepo.GetUsers(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	for i := range users {
		users[i].Roles, err = s.userRepo.GetUserRole(ctx, users[i].ID)
		if err != nil {
			logger.Error("s.userRepo.GetUserRole(): ", err)
		}
	}

	return users, total, nil
}

func (s *UserService) GetUser(ctx context.Context, filter *model.UserFilter, currentUserId int64) (*model.User, error) {

	if filter.Id == nil {
		return nil, model.ErrBadRequestQuery
	}

	role, err := s.userRepo.GetUserRole(ctx, currentUserId)
	if err != nil {
		return nil, err
	}

	if currentUserId != *filter.Id && !model.HasRoles(role, model.Role_Admin) {
		return nil, model.ErrAccessDenied
	}

	user, err := s.userRepo.GetUser(ctx, filter)
	if err != nil {
		return nil, err
	}

	user.Roles, err = s.userRepo.GetUserRole(ctx, user.ID)
	if err != nil {
		logger.Error("s.userRepo.GetUserRole(): ", err)
		return nil, err
	}

	return user, nil
}

func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*model.Token, error) {
	token, err := s.token.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	if token.RtExpires < time.Now().UTC().Unix() {
		return nil, errors.New("refresh token has expired")
	}

	err = s.verifyRefreshToken(ctx, token.UserId, token.RefreshUuid, token.SessionUuid)
	if err != nil {
		return nil, err
	}

	userId := token.UserId
	user, err := s.userRepo.GetUser(ctx, &model.UserFilter{Id: &userId})
	if err != nil {
		return nil, err
	}

	err = s.redis.Delete(ctx, token.RefreshUuid, token.SessionUuid)
	if err != nil {
		return nil, err
	}

	return s.createSession(ctx, user.ID)
}

func (s *UserService) VerifyAccessToken(ctx context.Context, userId int64, accessUUID, sessionUuid string) error {
	encoded, err := s.redis.Get(ctx, accessUUID)
	if err != nil {
		return err
	}

	var accessPayload []string

	err = json.Unmarshal([]byte(encoded), &accessPayload)
	if err != nil {
		return err
	}

	if len(accessPayload) != 2 || accessPayload[0] != sessionUuid {
		return errors.New("token is not verified")
	}

	encoded, err = s.redis.Get(ctx, sessionUuid)
	if err != nil {
		return err
	}

	var encodeUserId int64
	err = json.Unmarshal([]byte(encoded), &encodeUserId)
	if err != nil {
		return err
	}

	if encodeUserId != userId {
		return errors.New("token is not verified")
	}

	return nil
}

func (s *UserService) Logout(ctx context.Context, token *model.TokenDetails) error {
	err := s.VerifyAccessToken(ctx, token.UserId, token.AccessUuid, token.SessionUuid)
	if err != nil {
		return err
	}

	err = s.redis.Delete(ctx, token.AccessUuid, token.SessionUuid, token.RefreshUuid)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, adminUserID, userID int64) error {

	role, err := s.userRepo.GetUserRole(ctx, adminUserID)
	if err != nil {
		return err
	}

	if !model.HasRoles(role, model.Role_Admin) {
		return model.ErrAccessDenied
	}

	err = s.userRepo.DeleteUser(ctx, userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) verifyRefreshToken(ctx context.Context, userId int64, refreshUuid, sessionUuid string) error {
	encoded, err := s.redis.Get(ctx, refreshUuid)
	if err != nil {
		return err
	}

	var refreshPayload string

	err = json.Unmarshal([]byte(encoded), &refreshPayload)
	if err != nil {
		return err
	}

	if refreshPayload != sessionUuid {
		return errors.New("token is not verified")
	}

	encoded, err = s.redis.Get(ctx, sessionUuid)
	if err != nil {
		return err
	}

	var encodeUserId int64
	err = json.Unmarshal([]byte(encoded), &encodeUserId)
	if err != nil {
		return err
	}

	if encodeUserId != userId {
		return errors.New("token is not verified")
	}

	return nil
}

func (s *UserService) createSession(ctx context.Context, userId int64) (*model.Token, error) {
	if userId == 0 {
		return nil, model.ErrSignIn
	}

	tokens, err := s.token.NewTokenDetails(userId)
	if err != nil {
		logger.Error("NewTokenDetails(): ", err)
		return nil, err
	}

	err = s.redis.Save(ctx, tokens.SessionUuid, userId, time.Duration(tokens.RtExpires)*time.Second)
	if err != nil {
		logger.Error("Save(): ", err)
		return nil, err
	}

	err = s.saveTokens(ctx, tokens)
	if err != nil {
		logger.Error("saveTokens(): ", err)
		return nil, err
	}

	return &model.Token{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (s *UserService) saveTokens(ctx context.Context, tokenDetail *model.TokenDetails) error {
	err := s.redis.Save(ctx, tokenDetail.AccessUuid, []string{tokenDetail.SessionUuid, tokenDetail.RefreshUuid}, time.Duration(tokenDetail.AtExpires)*time.Second)
	if err != nil {
		return err
	}

	err = s.redis.Save(ctx, tokenDetail.RefreshUuid, tokenDetail.SessionUuid, time.Duration(tokenDetail.RtExpires)*time.Second)
	if err != nil {
		return err
	}

	return nil
}

func (s *UserService) checkUsernameAndEmailExists(ctx context.Context, username, email string) error {

	if exists, _ := s.userRepo.UsernameExists(ctx, username); exists {
		return model.ErrUsernameAlreadyExists
	}

	if exists, _ := s.userRepo.EmailExists(ctx, email); exists {
		return model.ErrEmailAlreadyExists
	}

	return nil
}
