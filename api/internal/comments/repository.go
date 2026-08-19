package comments

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Comment struct {
	ID              string    `json:"id"`
	PostID          string    `json:"post_id"`
	AuthorID        string    `json:"author_id"`
	Username        string    `json:"username"`
	DisplayName     string    `json:"display_name"`
	Content         string    `json:"content"`
	ParentCommentID *string   `json:"parent_comment_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
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

func (repo *Repository) Create(
	ctx context.Context,
	userID string,
	postID string,
	content string,
) (*Comment, error) {
	if content == "" {
		return nil, fmt.Errorf("comment cannot be empty")
	}

	allowed, err := repo.canAccessPost(ctx, userID, postID)

	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, fmt.Errorf("you cannot access this post")
	}

	comment := &Comment{}

	err = repo.db.QueryRow(
		ctx,
		`
		INSERT INTO comments (
			post_id,
			author_id,
			content
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			post_id,
			author_id,
			content,
			parent_comment_id,
			created_at
		`,
		postID,
		userID,
		content,
	).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.AuthorID,
		&comment.Content,
		&comment.ParentCommentID,
		&comment.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (repo *Repository) GetForPost(
	ctx context.Context,
	userID string,
	postID string,
) ([]Comment, error) {
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
		SELECT
			c.id,
			c.post_id,
			c.author_id,
			u.username,
			u.display_name,
			c.content,
			c.parent_comment_id,
			c.created_at
		FROM comments c
		JOIN users u
			ON u.id = c.author_id
		WHERE c.post_id = $1
		  AND c.status = 'active'
		ORDER BY c.created_at ASC
		`,
		postID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	result := make([]Comment, 0)

	for rows.Next() {
		var comment Comment

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.AuthorID,
			&comment.Username,
			&comment.DisplayName,
			&comment.Content,
			&comment.ParentCommentID,
			&comment.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		result = append(result, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
