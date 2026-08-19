package reactions

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Reaction struct {
	Type string `json:"type"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) canAccessPost(
	ctx context.Context,
	userID string,
	postID string,
) (bool, error) {
	var allowed bool

	err := repo.db.QueryRow(
		ctx,
		`
		SELECT EXISTS (
			SELECT 1
			FROM posts p
			WHERE p.id = $1
			  AND p.status = 'active'
			  AND (
				p.author_id = $2
				OR p.visibility = 'public'
				OR (
					p.visibility = 'inner_circle'
					AND EXISTS (
						SELECT 1
						FROM inner_circle_members ic
						WHERE ic.user_id = $2
						  AND ic.member_id = p.author_id
					)
				)
			  )
		)
		`,
		postID,
		userID,
	).Scan(&allowed)

	return allowed, err
}

func isValidReaction(reactionType string) bool {
	switch reactionType {
	case "like", "support", "helpful":
		return true
	default:
		return false
	}
}

func (repo *Repository) Add(
	ctx context.Context,
	userID string,
	postID string,
	reactionType string,
) error {
	if !isValidReaction(reactionType) {
		return fmt.Errorf("invalid reaction type")
	}

	allowed, err := repo.canAccessPost(ctx, userID, postID)

	if err != nil {
		return err
	}

	if !allowed {
		return fmt.Errorf("you cannot access this post")
	}

	_, err = repo.db.Exec(
		ctx,
		`
		INSERT INTO post_reactions (
			post_id,
			user_id,
			reaction_type
		)
		VALUES ($1, $2, $3)
		ON CONFLICT (post_id, user_id, reaction_type)
		DO NOTHING
		`,
		postID,
		userID,
		reactionType,
	)

	return err
}

func (repo *Repository) Remove(
	ctx context.Context,
	userID string,
	postID string,
	reactionType string,
) error {
	if !isValidReaction(reactionType) {
		return fmt.Errorf("invalid reaction type")
	}

	_, err := repo.db.Exec(
		ctx,
		`
		DELETE FROM post_reactions
		WHERE post_id = $1
		  AND user_id = $2
		  AND reaction_type = $3
		`,
		postID,
		userID,
		reactionType,
	)

	return err
}

func (repo *Repository) GetMyReactions(
	ctx context.Context,
	userID string,
	postID string,
) ([]Reaction, error) {
	allowed, err := repo.canAccessPost(ctx, userID, postID)

	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, fmt.Errorf("you cannot access this post")
	}

	rows, err := repo.db.Query(
		ctx,
		`
		SELECT reaction_type
		FROM post_reactions
		WHERE post_id = $1
		  AND user_id = $2
		ORDER BY created_at ASC
		`,
		postID,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	reactions := make([]Reaction, 0)

	for rows.Next() {
		var reaction Reaction

		if err := rows.Scan(&reaction.Type); err != nil {
			return nil, err
		}

		reactions = append(reactions, reaction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reactions, nil
}
