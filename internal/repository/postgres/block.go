package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/model"
)

type IBlockRepo interface {
	Create(ctx context.Context, block *model.Block) error
	GetBlock(ctx context.Context, id int64) (*model.Block, error)
	GetBlocks(ctx context.Context, pageID int64) ([]*model.Block, error)
	Update(ctx context.Context, block *model.Block) error
	Delete(ctx context.Context, id int64) error
	GetEmails(ctx context.Context) ([]string, error)
}

type BlockRepo struct {
	db *sql.DB
}

func NewBlockRepo(db *sql.DB) *BlockRepo {
	return &BlockRepo{db: db}
}

func (p *BlockRepo) Create(ctx context.Context, block *model.Block) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	err = tx.QueryRowContext(ctx, `
		INSERT INTO public.block (page_id, block_type_id)
		VALUES ($1, $2)
		RETURNING id
	`, block.PageID, block.Type.ID).Scan(&block.ID)
	if err != nil {
		return fmt.Errorf("insert block: %w", err)
	}

	for _, content := range block.BlockContents {
		err = tx.QueryRowContext(ctx, `
			INSERT INTO public.block_content (block_id, created_at, updated_at, image_url)
			VALUES ($1, now(), now(), $2)
			RETURNING id
		`, block.ID, content.ImageURL).Scan(&content.ID)
		if err != nil {
			return fmt.Errorf("insert block_content: %w", err)
		}

		for code, tr := range content.Translation {
			langID, ok := model.LanguageMap[code]
			if !ok {
				return fmt.Errorf("unknown language code: %s", code)
			}

			_, err = tx.ExecContext(ctx, `
				INSERT INTO public.block_content_translations (block_content_id, language_id, title, body)
				VALUES ($1, $2, $3, $4)
			`, content.ID, langID, tr.Name, tr.Description)
			if err != nil {
				return fmt.Errorf("insert block_content_translations: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit block: %w", err)
	}

	return nil
}

func (p *BlockRepo) GetBlock(ctx context.Context, id int64) (*model.Block, error) {
	block := &model.Block{
		Translation:   make(map[string]model.BlockTranslation),
		BlockContents: []*model.BlockContent{},
	}

	// 🔹 1. Основной блок + тип
	err := p.db.QueryRowContext(ctx, `
		SELECT b.id, b.page_id, bt.id, bt.code, bt.name
		FROM public.block b
		LEFT JOIN public.block_type bt ON bt.id = b.block_type_id
		WHERE b.id = $1 AND b.is_active = true
		ORDER BY b.id asc
	`, id).Scan(
		&block.ID,
		&block.PageID,
		&block.Type.ID,
		&block.Type.Code,
		&block.Type.Name,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("select block: %w", err)
	}

	// 🔹 2. Переводы блока (если появится таблица block_translations — пока просто пропускаем)

	// 🔹 3. Контент блоков
	contentRows, err := p.db.QueryContext(ctx, `
		SELECT id, created_at, updated_at, image_url
		FROM public.block_content
		WHERE block_id = $1
		ORDER BY id asc
	`, id)
	if err != nil {
		return nil, fmt.Errorf("select block_content: %w", err)
	}
	defer func() {
		_ = contentRows.Close()
	}()

	for contentRows.Next() {
		content := &model.BlockContent{
			Translation: make(map[string]model.BlockTranslation),
		}
		if err := contentRows.Scan(&content.ID, &content.CreatedAt, &content.UpdatedAt, &content.ImageURL); err != nil {
			return nil, err
		}

		// 🔹 4. Переводы контента (убрали поля created_at/updated_at)
		trRows, err := p.db.QueryContext(ctx, `
			SELECT bct.id, l.code, bct.title, bct.body
			FROM public.block_content_translations bct
			INNER JOIN public.language l ON l.id = bct.language_id
			WHERE bct.block_content_id = $1
			order by bct.id asc
		`, content.ID)
		if err != nil {
			return nil, fmt.Errorf("select block_content_translations: %w", err)
		}

		for trRows.Next() {
			var tr model.BlockTranslation
			var langCode string
			if err := trRows.Scan(&tr.ID, &langCode, &tr.Name, &tr.Description); err != nil {
				return nil, err
			}
			content.Translation[langCode] = tr
		}
		if err := trRows.Err(); err != nil {
			_ = trRows.Close()
			return nil, err
		}
		if err := trRows.Close(); err != nil {
			return nil, err
		}

		block.BlockContents = append(block.BlockContents, content)
	}

	return block, nil
}

func (p *BlockRepo) GetBlocks(ctx context.Context, pageID int64) ([]*model.Block, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT b.id
		FROM public.block b
		WHERE b.page_id = $1 AND b.is_deleted = false
		ORDER BY b.id asc
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var blocks []*model.Block
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}

		block, err := p.GetBlock(ctx, id)
		if err != nil {
			return nil, err
		}
		if block != nil {
			blocks = append(blocks, block)
		}
	}

	return blocks, nil
}

func (p *BlockRepo) Update(ctx context.Context, block *model.Block) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx, `
		UPDATE public.block
		SET block_type_id = $2, updated_at = now()
		WHERE id = $1 AND is_deleted = false
	`, block.ID, block.Type.ID)
	if err != nil {
		return fmt.Errorf("update block: %w", err)
	}

	for _, content := range block.BlockContents {
		var exists bool
		err = tx.QueryRowContext(ctx, `
			SELECT EXISTS(SELECT 1 FROM public.block_content WHERE id = $1 AND block_id = $2)
		`, content.ID, block.ID).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check block_content exists: %w", err)
		}

		if exists {
			_, err = tx.ExecContext(ctx, `
				UPDATE public.block_content
				SET updated_at = now(), image_url = coalesce($2, image_url)
				WHERE id = $1
			`, content.ID, content.ImageURL)
			if err != nil {
				return fmt.Errorf("update block_content: %w", err)
			}
		} else {
			err = tx.QueryRowContext(ctx, `
				INSERT INTO public.block_content (block_id, created_at, updated_at)
				VALUES ($1, now(), now())
				RETURNING id
			`, block.ID).Scan(&content.ID)
			if err != nil {
				return fmt.Errorf("insert new block_content: %w", err)
			}
		}

		for code, tr := range content.Translation {
			langID, ok := model.LanguageMap[code]
			if !ok {
				return fmt.Errorf("unknown language code: %s", code)
			}

			var trExists bool
			err = tx.QueryRowContext(ctx, `
				SELECT EXISTS(
					SELECT 1 FROM public.block_content_translations 
					WHERE block_content_id = $1 AND language_id = $2
				)
			`, content.ID, langID).Scan(&trExists)
			if err != nil {
				return fmt.Errorf("check translation exists: %w", err)
			}

			if trExists {
				_, err = tx.ExecContext(ctx, `
					UPDATE public.block_content_translations
					SET title = $3, body = $4, updated_at = now()
					WHERE block_content_id = $1 AND language_id = $2
				`, content.ID, langID, tr.Name, tr.Description)
				if err != nil {
					return fmt.Errorf("update translation: %w", err)
				}
			} else {
				_, err = tx.ExecContext(ctx, `
					INSERT INTO public.block_content_translations (block_content_id, language_id, title, body, created_at, updated_at)
					VALUES ($1, $2, $3, $4, now(), now())
				`, content.ID, langID, tr.Name, tr.Description)
				if err != nil {
					return fmt.Errorf("insert new translation: %w", err)
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit block update: %w", err)
	}

	return nil
}

func (p *BlockRepo) Delete(ctx context.Context, id int64) error {
	_, err := p.db.ExecContext(ctx, `
		UPDATE public.block
		SET is_deleted = true, updated_at = now()
		WHERE id = $1
	`, id)
	return err
}

func (p *BlockRepo) GetEmails(ctx context.Context) ([]string, error) {
	const fixedBlockTypeID = 14

	rows, err := p.db.QueryContext(ctx, `
        SELECT DISTINCT bct.title
        FROM public.block_content_translations bct
        INNER JOIN public.block_content bc ON bc.id = bct.block_content_id
        INNER JOIN public.block b ON b.id = bc.block_id
        WHERE b.block_type_id = $1
        ORDER BY bct.title ASC
    `, fixedBlockTypeID)
	if err != nil {
		return nil, fmt.Errorf("select emails: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, fmt.Errorf("scan email: %w", err)
		}
		emails = append(emails, email)
	}

	return emails, nil
}
