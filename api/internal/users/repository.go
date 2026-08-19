package users

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Username      string `json:"username"`
	DisplayName   string `json:"display_name"`
	Bio           string `json:"bio"`
	AvatarURL     string `json:"avatar_url"`
	HumanVerified bool   `json:"human_verified"`
	Status        string `json:"status"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) GetByID(ctx context.Context, userID string) (*User, error) {
	user := &User{}

	err := repo.db.QueryRow(
		ctx,
		`
		SELECT
			id,
			email,
			username,
			display_name,
			COALESCE(bio, ''),
			COALESCE(avatar_url, ''),
			human_verified,
			status
		FROM users
		WHERE id = $1
		`,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.DisplayName,
		&user.Bio,
		&user.AvatarURL,
		&user.HumanVerified,
		&user.Status,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (repo *Repository) UpdateProfile(
	ctx context.Context,
	userID string,
	displayName string,
	bio string,
	avatarURL string,
) (*User, error) {
	user := &User{}

	err := repo.db.QueryRow(
		ctx,
		`
		UPDATE users
		SET
			display_name = $1,
			bio = $2,
			avatar_url = $3,
			updated_at = NOW()
		WHERE id = $4
		RETURNING
			id,
			email,
			username,
			display_name,
			COALESCE(bio, ''),
			COALESCE(avatar_url, ''),
			human_verified,
			status
		`,
		displayName,
		bio,
		avatarURL,
		userID,
	).Scan(
		&user.ID,
		&user.Email,
		&user.Username,
		&user.DisplayName,
		&user.Bio,
		&user.AvatarURL,
		&user.HumanVerified,
		&user.Status,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}