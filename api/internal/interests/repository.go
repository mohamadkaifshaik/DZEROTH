package interests

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Interest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) GetAll(
	ctx context.Context,
) ([]Interest, error) {
	rows, err := repo.db.Query(
		ctx,
		`
		SELECT
			id,
			name,
			slug,
			COALESCE(description, '')
		FROM interests
		ORDER BY name ASC
		`,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]Interest, 0)

	for rows.Next() {
		var interest Interest

		err := rows.Scan(
			&interest.ID,
			&interest.Name,
			&interest.Slug,
			&interest.Description,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, interest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) GetForUser(
	ctx context.Context,
	userID string,
) ([]Interest, error) {
	rows, err := repo.db.Query(
		ctx,
		`
		SELECT
			i.id,
			i.name,
			i.slug,
			COALESCE(i.description, '')
		FROM interests i
		JOIN user_interests ui
			ON ui.interest_id = i.id
		WHERE ui.user_id = $1
		ORDER BY i.name ASC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]Interest, 0)

	for rows.Next() {
		var interest Interest

		err := rows.Scan(
			&interest.ID,
			&interest.Name,
			&interest.Slug,
			&interest.Description,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, interest)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (repo *Repository) AddForUser(
	ctx context.Context,
	userID string,
	interestID string,
) error {
	_, err := repo.db.Exec(
		ctx,
		`
		INSERT INTO user_interests (
			user_id,
			interest_id
		)
		VALUES ($1, $2)
		ON CONFLICT (user_id, interest_id)
		DO NOTHING
		`,
		userID,
		interestID,
	)

	return err
}

func (repo *Repository) RemoveForUser(
	ctx context.Context,
	userID string,
	interestID string,
) error {
	_, err := repo.db.Exec(
		ctx,
		`
		DELETE FROM user_interests
		WHERE user_id = $1
		  AND interest_id = $2
		`,
		userID,
		interestID,
	)

	return err
}
