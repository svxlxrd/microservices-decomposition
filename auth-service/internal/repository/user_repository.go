package repository

import (
	"bookshelf/auth-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	user.ID = uuid.New().String()

	query := `
	INSERT INTO users (id, username, email, password_hash)
	VALUES (:id, :username, :email, :password_hash)
	RETURNING created_at, updated_at;`

	rows, err := r.db.NamedQueryContext(ctx, query, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan timestamps: %w", err)
		}
	}

	return &domain.User{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
	}, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
	SELECT * FROM users
	WHERE id=$1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
	SELECT * FROM users
	WHERE email=$1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
	SELECT * FROM users
	WHERE username=$1`

	var user domain.User
	err := r.db.GetContext(ctx, &user, query, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
	UPDATE users
	SET username = :username, updated_at = NOW()
	WHERE id = :id`

	result, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	err := r.db.GetContext(ctx, &exists, query, email)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) UsernameExists(ctx context.Context, username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

	err := r.db.GetContext(ctx, &exists, query, username)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) GetByIDs(ctx context.Context, ids []string) (map[string]*domain.User, error) {
	if len(ids) == 0 {
		return make(map[string]*domain.User), nil
	}

	query := `
		SELECT id, username, email, created_at, updated_at 
		FROM users 
		WHERE id = ANY($1)`

	var users []*domain.User
	err := r.db.SelectContext(ctx, &users, query, pq.Array(ids))
	if err != nil {
		return nil, fmt.Errorf("failed to get users by ids: %w", err)
	}

	result := make(map[string]*domain.User, len(users))
	for _, u := range users {
		result[u.ID] = u
	}

	return result, nil
}
