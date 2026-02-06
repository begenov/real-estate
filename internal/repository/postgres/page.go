package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/model"
	"time"
)

type IPageRepo interface {
	Create(ctx context.Context, page *model.Page) error
	Update(ctx context.Context, page *model.Page) error
	GetPages(ctx context.Context) ([]*model.Page, error)
	GetPage(ctx context.Context, id int64) (*model.Page, error)
	Delete(ctx context.Context, id int64) error
}

type PageRepo struct {
	db *sql.DB
}

func NewPageRepo(db *sql.DB) *PageRepo {
	return &PageRepo{
		db: db,
	}
}

func (p *PageRepo) GetPages(ctx context.Context) ([]*model.Page, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			p.id,
			p.slug,
			p.created_at,
			p.updated_at,
			pt.id,
			pt.title,
			pt.description,
			l.id,
			l.code,
			pt.created_at,
			pt.updated_at
		FROM public.page p
		LEFT JOIN public.page_translations pt ON pt.page_id = p.id
		LEFT JOIN public."language" l ON l.id = pt.language_id
		WHERE p.is_deleted = false
		ORDER BY p.id asc
	`)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	pagesMap := make(map[int64]*model.Page)
	var pagesOrder = make([]int64, 0)

	for rows.Next() {
		var (
			pageID                   sql.NullInt64
			pageSlug                 sql.NullString
			pageCreated, pageUpdated sql.NullTime

			trID                 sql.NullInt64
			trTitle              sql.NullString
			trDesc               sql.NullString
			langID               sql.NullInt64
			langCode             sql.NullString
			trCreated, trUpdated sql.NullTime
		)

		err := rows.Scan(
			&pageID,
			&pageSlug,
			&pageCreated,
			&pageUpdated,
			&trID,
			&trTitle,
			&trDesc,
			&langID,
			&langCode,
			&trCreated,
			&trUpdated,
		)
		if err != nil {
			return nil, err
		}

		page, ok := pagesMap[pageID.Int64]
		if !ok {
			pagesOrder = append(pagesOrder, pageID.Int64)
			page = &model.Page{
				ID:          pageID.Int64,
				Slug:        pageSlug.String,
				CreatedAt:   nilIfZero(pageCreated),
				UpdatedAt:   nilIfZero(pageUpdated),
				Translation: make(map[string]*model.PageTranslation),
			}
			pagesMap[pageID.Int64] = page
		}

		if trID.Valid && langCode.Valid {
			tr := &model.PageTranslation{
				ID:          trID.Int64,
				Name:        trTitle.String,
				Description: nullableString(trDesc),
				Language: model.Language{
					ID:   int(langID.Int64),
					Code: langCode.String,
				},
				CreatedAt: nilIfZero(trCreated),
				UpdatedAt: nilIfZero(trUpdated),
			}
			page.Translation[langCode.String] = tr
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	pages := make([]*model.Page, 0, len(pagesMap))
	for _, p := range pagesOrder {
		pages = append(pages, pagesMap[p])
	}

	return pages, nil
}

func (p *PageRepo) GetPage(ctx context.Context, id int64) (*model.Page, error) {
	page := &model.Page{}

	err := p.db.QueryRowContext(ctx, `
		SELECT id, slug, created_at, updated_at
		FROM public.page
		WHERE id = $1 and is_deleted = false
	`, id).Scan(
		&page.ID,
		&page.Slug,
		&page.CreatedAt,
		&page.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	rows, err := p.db.QueryContext(ctx, `
		SELECT 
			pt.id,
			pt.title,
			pt.description,
			pt.language_id,
			l.code,
			pt.created_at,
			pt.updated_at
		FROM public.page_translations pt
		INNER JOIN public."language" l ON l.id = pt.language_id
		WHERE pt.page_id = $1
	`, id)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	page.Translation = make(map[string]*model.PageTranslation)

	for rows.Next() {
		tr := &model.PageTranslation{Language: model.Language{}}
		err := rows.Scan(
			&tr.ID,
			&tr.Name,
			&tr.Description,
			&tr.Language.ID,
			&tr.Language.Code,
			&tr.CreatedAt,
			&tr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		page.Translation[tr.Language.Code] = tr
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return page, nil
}

func (p *PageRepo) Create(ctx context.Context, page *model.Page) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	query := `
		INSERT INTO public.page (slug, is_active, created_at, updated_at)
		VALUES ($1, true, now(), now())
		RETURNING id, created_at, updated_at;
	`
	err = tx.QueryRowContext(ctx, query, page.Slug).Scan(&page.ID, &page.CreatedAt, &page.UpdatedAt)
	if err != nil {
		return err
	}

	queryTr := `
		INSERT INTO public.page_translations (page_id, language_id, title, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`

	for _, tr := range page.Translation {
		err := tx.QueryRowContext(ctx, queryTr,
			page.ID,
			tr.Language.ID,
			tr.Name,
			tr.Description,
		).Scan(&tr.ID)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *PageRepo) Update(ctx context.Context, page *model.Page) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE public.page
		SET slug = $1,
		    updated_at = now()
		WHERE id = $2
	`, page.Slug, page.ID)
	if err != nil {
		return err
	}

	for _, tr := range page.Translation {
		var existingID int64

		err := tx.QueryRowContext(ctx, `
			SELECT id FROM public.page_translations
			WHERE page_id = $1 AND language_id = $2
		`, page.ID, tr.Language.ID).Scan(&existingID)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				err = tx.QueryRowContext(ctx, `
					INSERT INTO public.page_translations (page_id, language_id, title, description, updated_at)
					VALUES ($1, $2, $3, $4, now())
					RETURNING id
				`, page.ID, tr.Language.ID, tr.Name, tr.Description).Scan(&tr.ID)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			_, err = tx.ExecContext(ctx, `
				UPDATE public.page_translations
				SET title = $1,
				    description = $2,
				    updated_at = now()
				WHERE id = $3
			`, tr.Name, tr.Description, existingID)
			if err != nil {
				return err
			}
			tr.ID = existingID
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (p *PageRepo) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE public.page
		SET is_deleted = true,
		    updated_at = now()
		WHERE id = $1 AND is_deleted = false
		RETURNING id;
	`

	var deletedID int64
	err := p.db.QueryRowContext(ctx, query, id).Scan(&deletedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("page with id %d not found or already deleted", id)
		}
		return err
	}

	return nil
}

func nullableString(ns sql.NullString) *string {
	if ns.Valid {
		return &ns.String
	}
	return nil
}

func nilIfZero(nt sql.NullTime) *time.Time {
	if nt.Valid {
		return &nt.Time
	}
	return nil
}
