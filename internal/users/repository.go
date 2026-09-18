package users

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateUser(
	ctx context.Context,
	db *pgxpool.Pool,
	email string,
	passwordHash string,
) (User, error) {
	var user User

	err := db.QueryRow(
		ctx,
		`
		INSERT INTO users (
			email,
			password_hash
		)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at
		`,
		email,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func GetUserByEmail(
	ctx context.Context,
	db *pgxpool.Pool,
	email string,
) (User, error) {
	var user User

	err := db.QueryRow(
		ctx,
		`
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1
		`,
		email,
	).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)

	if err != nil {
		return User{}, fmt.Errorf("get user: %w", err)
	}

	return user, nil
}
