package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/begenov/real-estate/internal/model"
	"strings"
	"time"
)

type IUserRepo interface {
	GetUser(ctx context.Context, filter *model.UserFilter) (*model.User, error)
	UsernameExists(ctx context.Context, username string) (bool, error)
	EmailExists(ctx context.Context, email string) (bool, error)
	CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) error
	GetUserRole(ctx context.Context, userId int64) ([]model.Role, error)
	InsertUserRole(ctx context.Context, tx *sql.Tx, userID int64, roleIDs []int64) error
	GetUsers(ctx context.Context, filter *model.UsersFilter) ([]*model.User, int, error)
	UpdateUser(ctx context.Context, tx *sql.Tx, user *model.User) error
	DeleteUser(ctx context.Context, userID int64) error
}

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) IUserRepo {
	return &UserRepo{
		db: db,
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, tx *sql.Tx, user *model.User) error {
	query := `
		INSERT INTO public."user" (username, email, phone, first_name, last_name, middle_name, password, created_at, updated_at, owner_id, photo_url)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id;
	`

	var userID int64
	err := tx.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.Phone,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Password,
		time.Now(),
		time.Now(),
		user.OwnerId,
		user.PhotoURL,
	).Scan(&userID)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	user.ID = userID

	return nil
}

func (r *UserRepo) UpdateUser(ctx context.Context, tx *sql.Tx, user *model.User) error {
	query := `
		UPDATE public."user"
		SET 
			username = $1,
			email = $2,
			first_name = $3,
			last_name = $4,
			middle_name = $5,
			phone = $6,
			photo_url = COALESCE($7, photo_url),
			updated_at = NOW(),
			password = COALESCE($8, password),
			is_active = $9
		WHERE id = $10
	`

	_, err := tx.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.FirstName,
		user.LastName,
		user.MiddleName,
		user.Phone,
		user.PhotoURL,
		user.Password,
		user.IsActive,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}

	return nil
}

func (r *UserRepo) DeleteUser(ctx context.Context, userID int64) error {
	query := `UPDATE public."user" SET is_deleted = TRUE WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *UserRepo) GetUser(ctx context.Context, filter *model.UserFilter) (*model.User, error) {
	if filter.Id != nil && filter.Username != nil && filter.Email != nil {
		return nil, errors.New("id, username and email cannot both be provided")
	}

	query := `SELECT id, username, email, phone, first_name, last_name, middle_name, created_at, updated_at, password, owner_id, photo_url, is_active
			  FROM public."user" WHERE 1=1 and is_deleted = false `

	var args []interface{}

	if filter.Username != nil {
		query += ` AND username = $` + fmt.Sprint(len(args)+1)
		args = append(args, *filter.Username)
	}

	if filter.Email != nil {
		query += ` AND email = $` + fmt.Sprint(len(args)+1)
		args = append(args, *filter.Email)
	}

	if filter.Id != nil {
		query += ` AND id = $` + fmt.Sprint(len(args)+1)
		args = append(args, *filter.Id)
	}

	var user model.User
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Phone,
		&user.FirstName,
		&user.LastName,
		&user.MiddleName,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Password,
		&user.OwnerId,
		&user.PhotoURL,
		&user.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (r *UserRepo) GetUsers(ctx context.Context, filter *model.UsersFilter) ([]*model.User, int, error) {
	args := []interface{}{}
	whereParts := []string{}
	joins := ""
	argIndex := 1

	whereParts = append(whereParts, "u.is_deleted = false")

	if filter.RoleId != nil {
		joins += `JOIN public.user_roles ur ON ur.user_id = u.id `
		whereParts = append(whereParts, fmt.Sprintf("ur.role_id = $%d", argIndex))
		args = append(args, *filter.RoleId)
		argIndex++
	}

	if filter.UserId != nil {
		whereParts = append(whereParts, fmt.Sprintf("u.id = $%d", argIndex))
		args = append(args, *filter.UserId)
		argIndex++
	}

	whereClause := ""
	if len(whereParts) > 0 {
		whereClause = "WHERE " + strings.Join(whereParts, " AND ")
	}

	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM public."user" u %s %s`, joins, whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting users: %w", err)
	}

	limit := 10
	offset := 0
	if filter.Rows > 0 {
		limit = filter.Rows
	}
	if filter.Page > 1 {
		offset = (filter.Page - 1) * limit
	}

	args = append(args, limit, offset)

	query := fmt.Sprintf(`
		SELECT 
			u.id, u.username, u.email, u.first_name, u.last_name, u.middle_name,
			u.created_at, u.updated_at, u.phone, u.owner_id, u.photo_url, u.is_active
		FROM public."user" u
		%s
		%s
		ORDER BY u.id DESC
		LIMIT $%d OFFSET $%d
	`, joins, whereClause, argIndex, argIndex+1)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying users: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var users []*model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID,
			&u.Username,
			&u.Email,
			&u.FirstName,
			&u.LastName,
			&u.MiddleName,
			&u.CreatedAt,
			&u.UpdatedAt,
			&u.Phone,
			&u.OwnerId,
			&u.PhotoURL,
			&u.IsActive,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning user: %w", err)
		}
		users = append(users, &u)
	}

	return users, total, nil
}

func (r *UserRepo) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM public."user" WHERE username = $1 and is_deleted = false)`

	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking username: %w", err)
	}
	return exists, nil
}

func (r *UserRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM public."user" WHERE email = $1 and is_deleted = false)`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking email: %w", err)
	}
	return exists, nil
}

func (r *UserRepo) GetUserRole(ctx context.Context, userId int64) ([]model.Role, error) {
	query := `SELECT r.id, r.code 
              FROM public.user_roles ur
              JOIN public.role r ON ur.role_id = r.id
              WHERE ur.user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("error querying roles: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.ID, &role.Code); err != nil {
			return nil, fmt.Errorf("error scanning role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

func (r *UserRepo) InsertUserRole(ctx context.Context, tx *sql.Tx, userID int64, roleIDs []int64) error {
	query := `INSERT INTO public.user_roles (user_id, role_id) VALUES ($1, $2)`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		_ = stmt.Close()
	}()

	for _, roleID := range roleIDs {
		_, err := stmt.ExecContext(ctx, userID, roleID)
		if err != nil {
			return fmt.Errorf("failed to insert user role (userID: %d, roleID: %d): %w", userID, roleID, err)
		}
	}

	return nil
}
