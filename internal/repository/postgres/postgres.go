// Package postgres provides PostgreSQL-based implementations of
// repository interfaces used by the application.
// It is responsible for persisting and retrieving users and secrets
// using a relational database.
package postgres

import (
	"context"
	"database/sql"
	"errors"

	"gophkeeper/internal/model"
	"gophkeeper/internal/repository"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Storage implements UserRepository and SecretRepository
// using PostgreSQL as the underlying storage.
type Storage struct {
	db *sql.DB
}

// New creates a new PostgreSQL storage instance.
// The provided database connection is used for all repository operations.
func New(db *sql.DB) *Storage {
	return &Storage{db: db}
}

// Create inserts a new user into the database.
// The user ID and creation timestamp are populated after successful insertion.
func (s *Storage) Create(ctx context.Context, user *model.User) error {
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO users (login, password_hash)
		VALUES ($1, $2)
		RETURNING id, created_at
	`, user.Login, user.PasswordHash)
	return row.Scan(&user.ID, &user.CreatedAt)
}

// GetByLogin retrieves a user by login.
// It returns repository.ErrNotFound if the user does not exist.
func (s *Storage) GetByLogin(ctx context.Context, login string) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE login = $1
	`, login).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// GetByID retrieves a user by identifier.
// It returns repository.ErrNotFound if the user does not exist.
func (s *Storage) GetByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := s.db.QueryRowContext(ctx, `
		SELECT id, login, password_hash, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(&u.ID, &u.Login, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

// Upsert inserts or updates a secret identified by its ID and user ID.
// If the secret ID is zero, a new record is created.
// If no rows are affected during update, repository.ErrNotFound is returned.
func (s *Storage) Upsert(ctx context.Context, sec *model.Secret) (*model.Secret, error) {
	if sec.ID == 0 {
		row := s.db.QueryRowContext(ctx, `
			INSERT INTO secrets (user_id, type, title, meta, encrypted_data, version, updated_at, deleted)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
			RETURNING id
		`, sec.UserID, string(sec.Type), sec.Title, sec.Meta, sec.EncryptedData, sec.Version, sec.UpdatedAt, sec.Deleted)
		if err := row.Scan(&sec.ID); err != nil {
			return nil, err
		}
		return sec, nil
	}

	res, err := s.db.ExecContext(ctx, `
		UPDATE secrets
		SET type=$1, title=$2, meta=$3, encrypted_data=$4, version=$5, updated_at=$6, deleted=$7
		WHERE id=$8 AND user_id=$9
	`, string(sec.Type), sec.Title, sec.Meta, sec.EncryptedData, sec.Version, sec.UpdatedAt, sec.Deleted, sec.ID, sec.UserID)
	if err != nil {
		return nil, err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return nil, repository.ErrNotFound
	}
	return sec, nil
}

// List returns all secrets belonging to the specified user.
// Secrets are ordered by their identifier in ascending order.
func (s *Storage) List(ctx context.Context, userID int64) ([]model.Secret, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, type, title, meta, encrypted_data, version, updated_at, deleted
		FROM secrets
		WHERE user_id = $1
		ORDER BY id ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []model.Secret
	for rows.Next() {
		var sec model.Secret
		var typ string
		if err := rows.Scan(
			&sec.ID,
			&sec.UserID,
			&typ,
			&sec.Title,
			&sec.Meta,
			&sec.EncryptedData,
			&sec.Version,
			&sec.UpdatedAt,
			&sec.Deleted,
		); err != nil {
			return nil, err
		}
		sec.Type = model.SecretType(typ)
		out = append(out, sec)
	}
	return out, rows.Err()
}

// Get retrieves a secret by its identifier and user identifier.
// It returns repository.ErrNotFound if the secret does not exist.
func (s *Storage) Get(ctx context.Context, id int64, userID int64) (*model.Secret, error) {
	var sec model.Secret
	var typ string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, type, title, meta, encrypted_data, version, updated_at, deleted
		FROM secrets
		WHERE id = $1 AND user_id = $2
	`, id, userID).Scan(
		&sec.ID,
		&sec.UserID,
		&typ,
		&sec.Title,
		&sec.Meta,
		&sec.EncryptedData,
		&sec.Version,
		&sec.UpdatedAt,
		&sec.Deleted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	sec.Type = model.SecretType(typ)
	return &sec, nil
}

// Delete removes a secret identified by its ID and user ID.
// If the secret does not exist, repository.ErrNotFound is returned.
func (s *Storage) Delete(ctx context.Context, id int64, userID int64) error {
	res, err := s.db.ExecContext(ctx,
		`DELETE FROM secrets WHERE id=$1 AND user_id=$2`,
		id, userID,
	)
	if err != nil {
		return err
	}
	aff, _ := res.RowsAffected()
	if aff == 0 {
		return repository.ErrNotFound
	}
	return nil
}
