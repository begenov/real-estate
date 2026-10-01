package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/begenov/real-estate/internal/model"
)

// ── GET /api/v1/real_estate/public ───────────────────────────────────────────

func TestGetRealEstates_Success(t *testing.T) {
	reSvc := &mockRealEstateSvc{
		GetRealEstatesFn: func(_ context.Context, _ model.RealEstateFilter, _ *int64) (model.RealEstates, int, error) {
			return model.RealEstates{{ID: 1, Price: 100_000}}, 1, nil
		},
	}
	r := buildRouter(nil, reSvc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/real_estate/public", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetRealEstates_InvalidPriceMin(t *testing.T) {
	r := buildRouter(nil, &mockRealEstateSvc{}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/real_estate/public?price_min=notanumber", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

// ── GET /api/v1/real_estate/public/:id ───────────────────────────────────────

func TestGetRealEstateByID_Success(t *testing.T) {
	reSvc := &mockRealEstateSvc{
		GetRealEstateFn: func(_ context.Context, _ int64, _ *int64) (*model.RealEstate, error) {
			return &model.RealEstate{ID: 42, Price: 200_000}, nil
		},
	}
	r := buildRouter(nil, reSvc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/real_estate/public/42", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetRealEstateByID_InvalidID(t *testing.T) {
	r := buildRouter(nil, &mockRealEstateSvc{}, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/real_estate/public/notanid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestGetRealEstateByID_NotFound(t *testing.T) {
	reSvc := &mockRealEstateSvc{
		GetRealEstateFn: func(_ context.Context, _ int64, _ *int64) (*model.RealEstate, error) {
			return nil, model.ErrNotFound
		},
	}
	r := buildRouter(nil, reSvc, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/real_estate/public/99", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d: %s", w.Code, w.Body.String())
	}
}

// ── DELETE /api/v1/real_estate/private/:id (authenticated) ───────────────────

func TestDeleteRealEstate_NoAuth(t *testing.T) {
	r := buildRouter(nil, &mockRealEstateSvc{}, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/real_estate/private/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
}

func TestDeleteRealEstate_InvalidID(t *testing.T) {
	userSvc := alwaysVerifyUserSvc()
	r := buildRouter(userSvc, &mockRealEstateSvc{}, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/real_estate/private/notanid", nil)
	req.Header.Set("Authorization", testBearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestDeleteRealEstate_AccessDenied(t *testing.T) {
	userSvc := alwaysVerifyUserSvc()
	reSvc := &mockRealEstateSvc{
		DeleteRealEstateFn: func(_ context.Context, _, _ int64) error {
			return model.ErrAccessDenied
		},
	}
	r := buildRouter(userSvc, reSvc, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/real_estate/private/1", nil)
	req.Header.Set("Authorization", testBearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("want 403, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDeleteRealEstate_Success(t *testing.T) {
	userSvc := alwaysVerifyUserSvc()
	reSvc := &mockRealEstateSvc{
		DeleteRealEstateFn: func(_ context.Context, _, _ int64) error { return nil },
	}
	r := buildRouter(userSvc, reSvc, nil, alwaysValidTokenMgr())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/real_estate/private/1", nil)
	req.Header.Set("Authorization", testBearerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("want 204, got %d: %s", w.Code, w.Body.String())
	}
}

// ── GET /api/v1/amenity ───────────────────────────────────────────────────────

func TestGetAllAmenities_Success(t *testing.T) {
	amenitySvc := &mockAmenitySvc{
		GetAllFn: func(_ context.Context) ([]*model.Amenity, error) {
			return []*model.Amenity{{ID: 1, NameRU: "Бассейн"}}, nil
		},
	}
	r := buildRouter(nil, nil, amenitySvc, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/amenity", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}
}
