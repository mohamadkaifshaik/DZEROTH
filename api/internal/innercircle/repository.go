package innercircle

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Member struct {
	ID          string `json:"id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) GetMembers(
	ctx context.Context,
	userID string,
) ([]Member, error) {
	rows, err := repo.db.Query(
		ctx,
		`
		SELECT
			u.id,
			u.username,
			u.display_name,
			COALESCE(u.avatar_url, '')
		FROM inner_circle_members ic
		JOIN users u
			ON u.id = ic.member_id
		WHERE ic.user_id = $1
		  AND u.status = 'active'
		ORDER BY ic.created_at ASC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	members := make([]Member, 0)

	for rows.Next() {
		var member Member

		err := rows.Scan(
			&member.ID,
			&member.Username,
			&member.DisplayName,
			&member.AvatarURL,
		)

		if err != nil {
			return nil, err
		}

		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return members, nil
}

func (repo *Repository) AddMember(
	ctx context.Context,
	userID string,
	memberID string,
) error {
	tx, err := repo.db.Begin(ctx)

	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	if userID == memberID {
		return fmt.Errorf("you cannot add yourself")
	}

	var memberExists bool

	err = tx.QueryRow(
		ctx,
		`
		SELECT EXISTS(
			SELECT 1
			FROM users
			WHERE id = $1
			  AND status = 'active'
		)
		`,
		memberID,
	).Scan(&memberExists)

	if err != nil {
		return err
	}

	if !memberExists {
		return fmt.Errorf("user not found")
	}

	var memberCount int

	err = tx.QueryRow(
		ctx,
		`
		SELECT COUNT(*)
		FROM inner_circle_members
		WHERE user_id = $1
		`,
		userID,
	).Scan(&memberCount)

	if err != nil {
		return err
	}

	if memberCount >= 25 {
		return fmt.Errorf("inner circle can contain at most 25 people")
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO inner_circle_members (
			user_id,
			member_id
		)
		VALUES ($1, $2)
		ON CONFLICT (user_id, member_id) DO NOTHING
		`,
		userID,
		memberID,
	)

	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (repo *Repository) RemoveMember(
	ctx context.Context,
	userID string,
	memberID string,
) error {
	result, err := repo.db.Exec(
		ctx,
		`
		DELETE FROM inner_circle_members
		WHERE user_id = $1
		  AND member_id = $2
		`,
		userID,
		memberID,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("member not found in inner circle")
	}

	return nil
}
