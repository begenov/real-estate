package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/begenov/real-estate/internal/model"
)

type ITourRepo interface {
	Create(ctx context.Context, req *model.TourRequest) error
	GetAll(ctx context.Context, filter model.TourRequestFilter) ([]model.TourRequest, int, error)
	GetByID(ctx context.Context, id int64) (*model.TourRequest, error)
	UpdateStatus(ctx context.Context, id int64, statusID int) error
}

type TourRepo struct {
	db *sql.DB
}

func NewTourRepo(db *sql.DB) ITourRepo {
	return &TourRepo{db: db}
}

func (r *TourRepo) Create(ctx context.Context, req *model.TourRequest) error {
	query := `
		INSERT INTO public.tour_request (name, phone, desired_date, real_estate_id, status_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		req.Name, req.Phone, req.DesiredDate, req.RealEstateID, model.TourStatusNew,
	).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
}

func (r *TourRepo) GetAll(ctx context.Context, filter model.TourRequestFilter) ([]model.TourRequest, int, error) {
	args := []interface{}{}
	where := "WHERE 1=1"

	if filter.StatusID != nil {
		args = append(args, *filter.StatusID)
		where += fmt.Sprintf(" AND status_id = $%d", len(args))
	}

	var count int
	err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM public.tour_request `+where, args...).Scan(&count)
	if err != nil {
		return nil, 0, fmt.Errorf("count tour requests: %w", err)
	}

	if filter.Rows <= 0 {
		filter.Rows = 10
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	args = append(args, filter.Rows, (filter.Page-1)*filter.Rows)
	query := fmt.Sprintf(`
		SELECT id, name, phone, desired_date, real_estate_id, status_id, created_at, updated_at
		FROM public.tour_request %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, len(args)-1, len(args))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("get tour requests: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var requests []model.TourRequest
	for rows.Next() {
		var req model.TourRequest
		if err := rows.Scan(
			&req.ID, &req.Name, &req.Phone, &req.DesiredDate,
			&req.RealEstateID, &req.StatusID, &req.CreatedAt, &req.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan tour request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, count, nil
}

func (r *TourRepo) GetByID(ctx context.Context, id int64) (*model.TourRequest, error) {
	var req model.TourRequest
	err := r.db.QueryRowContext(ctx, `
		SELECT id, name, phone, desired_date, real_estate_id, status_id, created_at, updated_at
		FROM public.tour_request WHERE id = $1`, id).
		Scan(&req.ID, &req.Name, &req.Phone, &req.DesiredDate,
			&req.RealEstateID, &req.StatusID, &req.CreatedAt, &req.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("get tour request by id: %w", err)
	}
	return &req, nil
}

func (r *TourRepo) UpdateStatus(ctx context.Context, id int64, statusID int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE public.tour_request SET status_id = $1, updated_at = now() WHERE id = $2`,
		statusID, id)
	if err != nil {
		return fmt.Errorf("update tour request status: %w", err)
	}
	return nil
}
