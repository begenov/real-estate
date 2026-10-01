package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/begenov/real-estate/internal/model"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func buildRouter(
	userSvc *mockUserSvc,
	reSvc *mockRealEstateSvc,
	amenitySvc *mockAmenitySvc,
	tokenMgr *mockTokenMgr,
) *gin.Engine {
	h := NewHandler(
		reSvc, userSvc, nil, nil, tokenMgr,
		amenitySvc, nil, nil, nil, nil, nil,
	)
	r := gin.New()
	r.Use(gin.Recovery(), errorHandlerMiddleware)
	api := r.Group("/api")
	h.Init(api)
	return r
}

// ── POST /api/v1/user/sign-in ─────────────────────────────────────────────────

func TestUserSignIn_Success(t *testing.T) {
	userSvc := &mockUserSvc{
		SignInFn: func(_ context.Context, _, _ string) (*model.Token, error) {
			return &model.Token{AccessToken: "at", RefreshToken: "rt"}, nil
		},
	}
	r := buildRouter(userSvc, nil, nil, nil)

	body, _ := json.Marshal(map[string]string{"username": "alice", "password": "password123"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/sign-in", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.Token
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.AccessToken != "at" {
		t.Fatalf("want access_token=at, got %q", resp.AccessToken)
	}
}

func TestUserSignIn_InvalidJSON(t *testing.T) {
	r := buildRouter(&mockUserSvc{}, nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/sign-in", bytes.NewBufferString("{bad json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestUserSignIn_WrongCredentials(t *testing.T) {
	userSvc := &mockUserSvc{
		SignInFn: func(_ context.Context, _, _ string) (*model.Token, error) {
			return nil, model.ErrSignIn
		},
	}
	r := buildRouter(userSvc, nil, nil, nil)

	body, _ := json.Marshal(map[string]string{"username": "alice", "password": "wrongpass"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/sign-in", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d: %s", w.Code, w.Body.String())
	}
}

// ── POST /api/v1/user/auth/refresh ────────────────────────────────────────────

func TestUserRefresh_Success(t *testing.T) {
	userSvc := &mockUserSvc{
		RefreshTokenFn: func(_ context.Context, _ string) (*model.Token, error) {
			return &model.Token{AccessToken: "new-at", RefreshToken: "new-rt"}, nil
		},
	}
	r := buildRouter(userSvc, nil, nil, nil)

	body, _ := json.Marshal(map[string]string{"token": "old-refresh-token"})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/user/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ── GET /api/v1/user/info (authenticated) ─────────────────────────────────────

func TestGetUserInfo_NoAuth(t *testing.T) {
	r := buildRouter(&mockUserSvc{}, nil, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/info", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestGetUserInfo_Success(t *testing.T) {
	userSvc := alwaysVerifyUserSvc()
	userSvc.GetUserFn = func(_ context.Context, _ *model.UserFilter, _ int64) (*model.User, error) {
		u := "alice"
		return &model.User{ID: 1, Username: &u}, nil
	}
	r := buildRouter(userSvc, nil, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/user/info", nil)
	req.Header.Set("Authorization", testBearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

// ── DELETE /api/v1/user/manager/:id (authenticated) ──────────────────────────

func TestDeleteManager_NoAuth(t *testing.T) {
	r := buildRouter(&mockUserSvc{}, nil, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/manager/5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestDeleteManager_AccessDenied(t *testing.T) {
	userSvc := alwaysVerifyUserSvc()
	userSvc.DeleteUserFn = func(_ context.Context, _, _ int64) error {
		return model.ErrAccessDenied
	}
	r := buildRouter(userSvc, nil, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/user/manager/5", nil)
	req.Header.Set("Authorization", testBearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d: %s", w.Code, w.Body.String())
	}
}
