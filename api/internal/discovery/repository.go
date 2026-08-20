package discovery

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Post struct {
	ID          string `json:"id"`
	AuthorID    string `json:"author_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url"`
	Content     string `json:"content"`
	Visibility  string `json:"visibility"`
	CreatedAt   string `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (repo *Repository) GetFeed(
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
			p.created_at::text
		FROM posts p
		JOIN users u
			ON u.id = p.author_id
		WHERE p.visibility = 'public'
		  AND p.status = 'active'
		  AND EXISTS (
				SELECT 1
				FROM post_interests pi
				JOIN user_interests ui
					ON ui.interest_id = pi.interest_id
				WHERE pi.post_id = p.id
				  AND ui.user_id = $1
		  )
		ORDER BY p.created_at DESC
		`,
		userID,
	)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	posts := make([]Post, 0)

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

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
