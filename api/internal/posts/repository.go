package posts

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID          string    `json:"id"`
	AuthorID    string    `json:"author_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Content     string    `json:"content"`
	Visibility  string    `json:"visibility"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (repo *Repository) Create(
	ctx context.Context,
	authorID string,
	content string,
	visibility string,
	interestIDs []string,
) (*Post, error) {
	if content == "" {
		return nil, fmt.Errorf("post content cannot be empty")
	}

	if visibility != "public" && visibility != "inner_circle" {
		return nil, fmt.Errorf("invalid post visibility")
	}

	// Inner Circle posts don't need interests.
	if visibility == "inner_circle" {
		interestIDs = nil
	}

	tx, err := repo.db.Begin(ctx)

	if err != nil {
		return nil, err
	}

	defer tx.Rollback(ctx)

	post := &Post{}

	err = tx.QueryRow(
		ctx,
		`
		INSERT INTO posts (
			author_id,
			content,
			visibility
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			author_id,
			content,
			visibility,
			created_at
		`,
		authorID,
		content,
		visibility,
	).Scan(
		&post.ID,
		&post.AuthorID,
		&post.Content,
		&post.Visibility,
		&post.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	for _, interestID := range interestIDs {
		_, err = tx.Exec(
			ctx,
			`
			INSERT INTO post_interests (
				post_id,
				interest_id
			)
			VALUES ($1, $2)
			ON CONFLICT (post_id, interest_id)
			DO NOTHING
			`,
			post.ID,
			interestID,
		)

		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return post, nil
}

func (repo *Repository) GetInnerCircleFeed(
	ctx context.Context,
	userID string,
) ([]Post, error) {
	rows, err := repo.db.Query(
		ctx,
		`
		SELECT
			p.id,
			p.author_id,
			u.username,
			u.display_name,
			COALESCE(u.avatar_url, ''),
			p.content,
			p.visibility,
			p.created_at
		FROM posts p
		JOIN users u
			ON u.id = p.author_id
		WHERE p.status = 'active'
		  AND (
				p.author_id = $1
				OR
				(
					p.visibility = 'inner_circle'
					AND EXISTS (
						SELECT 1
						FROM inner_circle_members ic
						WHERE ic.user_id = $1
						  AND ic.member_id = p.author_id
					)
				)
		  )
		ORDER BY p.created_at DESC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	feed := make([]Post, 0)

	for rows.Next() {
		var post Post

		err := rows.Scan(
			&post.ID,
			&post.AuthorID,
			&post.Username,
			&post.DisplayName,
			&post.AvatarURL,
			&post.Content,
			&post.Visibility,
			&post.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		feed = append(feed, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return feed, nil
}

func (repo *Repository) AddInterests(
	ctx context.Context,
	postID string,
	interestIDs []string,
) error {
	for _, interestID := range interestIDs {
		_, err := repo.db.Exec(
			ctx,
			`
			INSERT INTO post_interests (
				post_id,
				interest_id
			)
			VALUES ($1, $2)
			ON CONFLICT (post_id, interest_id)
			DO NOTHING
			`,
			postID,
			interestID,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
