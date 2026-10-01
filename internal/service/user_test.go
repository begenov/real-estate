package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/begenov/real-estate/internal/model"
)

func newUserService(userRepo *mockUserRepo, redis *mockRedis, token *mockTokenManager, txRepo *mockTxRepo) IUserService {
	return NewUserService(
		&mockHash{
			GenerateFn: func(p string) (string, error) { return "hashed_" + p, nil },
			CompareFn:  func(h, p string) error { return nil },
		},
		token,
		userRepo,
		redis,
		txRepo,
	)
}

func fixedTokenDetails() *model.TokenDetails {
	return &model.TokenDetails{
		AccessToken:  "access",
		RefreshToken: "refresh",
		AccessUuid:   "access-uuid",
		RefreshUuid:  "refresh-uuid",
		SessionUuid:  "session-uuid",
		AtExpires:    time.Now().Add(time.Hour).Unix(),
		RtExpires:    time.Now().Add(24 * time.Hour).Unix(),
		UserId:       1,
	}
}

func alwaysSaveRedis() *mockRedis {
	return &mockRedis{
		SaveFn:   func(_ context.Context, _ string, _ interface{}, _ time.Duration) error { return nil },
		GetFn:    func(_ context.Context, _ string) (string, error) { return "", nil },
		DeleteFn: func(_ context.Context, _ ...string) error { return nil },
	}
}

func alwaysTokenManager() *mockTokenManager {
	return &mockTokenManager{
		NewDetailsFn:        func(userId int64) (*model.TokenDetails, error) { return fixedTokenDetails(), nil },
		ParseAccessTokenFn:  func(t string) (*model.TokenDetails, error) { return fixedTokenDetails(), nil },
		ParseRefreshTokenFn: func(t string) (*model.TokenDetails, error) { return fixedTokenDetails(), nil },
	}
}

// ── SignIn ────────────────────────────────────────────────────────────────────

func TestSignIn_ShortPassword(t *testing.T) {
	svc := newUserService(&mockUserRepo{}, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	_, err := svc.SignIn(context.Background(), "user", "short")
	if !errors.Is(err, model.ErrPasswordTooShort) {
		t.Fatalf("want ErrPasswordTooShort, got %v", err)
	}
}

func TestSignIn_InactiveUser(t *testing.T) {
	password := "password123"
	hashed := "hashed_" + password

	repo := &mockUserRepo{
		GetUserFn: func(_ context.Context, _ *model.UserFilter) (*model.User, error) {
			return &model.User{ID: 1, Password: &hashed, IsActive: false}, nil
		},
	}

	svc := newUserService(repo, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	_, err := svc.SignIn(context.Background(), "user", password)
	if !errors.Is(err, model.ErrUserNotActive) {
		t.Fatalf("want ErrUserNotActive, got %v", err)
	}
}

func TestSignIn_WrongPassword(t *testing.T) {
	hashed := "correct_hash"
	repo := &mockUserRepo{
		GetUserFn: func(_ context.Context, _ *model.UserFilter) (*model.User, error) {
			return &model.User{ID: 1, Password: &hashed, IsActive: true}, nil
		},
	}

	hashMock := &mockHash{
		CompareFn: func(h, p string) error { return errors.New("mismatch") },
	}

	svc := NewUserService(hashMock, alwaysTokenManager(), repo, alwaysSaveRedis(), &mockTxRepo{})

	_, err := svc.SignIn(context.Background(), "user", "wrongpassword")
	if !errors.Is(err, model.ErrSignIn) {
		t.Fatalf("want ErrSignIn, got %v", err)
	}
}

func TestSignIn_Success(t *testing.T) {
	password := "validpassword"
	hashed := "hashed_" + password

	repo := &mockUserRepo{
		GetUserFn: func(_ context.Context, _ *model.UserFilter) (*model.User, error) {
			return &model.User{ID: 1, Password: &hashed, IsActive: true}, nil
		},
	}

	svc := newUserService(repo, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	token, err := svc.SignIn(context.Background(), "user", password)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.AccessToken == "" || token.RefreshToken == "" {
		t.Fatal("expected non-empty tokens")
	}
}

func TestSignIn_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{
		GetUserFn: func(_ context.Context, _ *model.UserFilter) (*model.User, error) {
			return nil, sql.ErrNoRows
		},
	}

	svc := newUserService(repo, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	_, err := svc.SignIn(context.Background(), "nouser", "validpassword")
	if !errors.Is(err, model.ErrSignIn) {
		t.Fatalf("want ErrSignIn, got %v", err)
	}
}

// ── DeleteUser ────────────────────────────────────────────────────────────────

func TestDeleteUser_NotAdmin(t *testing.T) {
	repo := &mockUserRepo{
		GetUserRoleFn: func(_ context.Context, _ int64) ([]model.Role, error) {
			return []model.Role{{ID: model.Role_Manager_ID, Code: model.Role_Manager}}, nil
		},
	}

	svc := newUserService(repo, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	err := svc.DeleteUser(context.Background(), 1, 2)
	if !errors.Is(err, model.ErrAccessDenied) {
		t.Fatalf("want ErrAccessDenied, got %v", err)
	}
}

func TestDeleteUser_AdminSuccess(t *testing.T) {
	deleted := false
	repo := &mockUserRepo{
		GetUserRoleFn: func(_ context.Context, _ int64) ([]model.Role, error) {
			return []model.Role{{ID: model.Role_Admin_ID, Code: model.Role_Admin}}, nil
		},
		DeleteUserFn: func(_ context.Context, _ int64) error {
			deleted = true
			return nil
		},
	}

	svc := newUserService(repo, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	err := svc.DeleteUser(context.Background(), 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !deleted {
		t.Fatal("expected DeleteUser to be called")
	}
}

// ── CreateUser ────────────────────────────────────────────────────────────────

func TestCreateUser_ShortPassword(t *testing.T) {
	svc := newUserService(&mockUserRepo{}, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	inp := &model.UserCreateInput{Password: strPtr("short")}
	err := svc.CreateUser(context.Background(), inp)
	if !errors.Is(err, model.ErrPasswordTooShort) {
		t.Fatalf("want ErrPasswordTooShort, got %v", err)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	repo := &mockUserRepo{
		UsernameExistsFn: func(_ context.Context, _ string) (bool, error) { return true, nil },
		EmailExistsFn:    func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	svc := newUserService(repo, alwaysSaveRedis(), alwaysTokenManager(), &mockTxRepo{})

	inp := &model.UserCreateInput{
		Username:  "existinguser",
		Email:     "new@test.com",
		Phone:     "+70000000000",
		FirstName: "A",
		LastName:  "B",
		Password:  strPtr("validpassword123"),
	}
	err := svc.CreateUser(context.Background(), inp)
	if !errors.Is(err, model.ErrUsernameAlreadyExists) {
		t.Fatalf("want ErrUsernameAlreadyExists, got %v", err)
	}
}

func strPtr(s string) *string { return &s }
