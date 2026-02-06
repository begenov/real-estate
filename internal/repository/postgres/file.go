package postgres

import (
	"context"
	"database/sql"
)

type IFileRepo interface {
	Create(ctx context.Context, filePath string, userId int64) (int64, error)
	Delete(ctx context.Context, fileID int64) error
	ListURLs(ctx context.Context, limit, offset int) ([]string, error)
}

type FileRepo struct {
	db *sql.DB
}

func NewFileRepo(db *sql.DB) *FileRepo {
	return &FileRepo{
		db: db,
	}
}

func (r *FileRepo) Create(ctx context.Context, filePath string, userId int64) (int64, error) {
	query := `INSERT INTO public.files (url, user_id) VALUES ($1, $2) RETURNING id`

	var fileID int64
	err := r.db.QueryRowContext(ctx, query, filePath, userId).Scan(&fileID)
	if err != nil {
		return 0, err
	}

	return fileID, nil
}

func (r *FileRepo) Delete(ctx context.Context, fileID int64) error {
	query := `DELETE FROM public.files WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, fileID)
	if err != nil {
		return err
	}

	return nil
}

func (r *FileRepo) ListURLs(ctx context.Context, limit, offset int) ([]string, error) {
	query := `SELECT url FROM public.files ORDER BY id DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	urls := make([]string, 0, limit)
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}
