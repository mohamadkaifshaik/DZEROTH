package httpapi

import (
	"net/http"
	"strings"

	"github.com/mohamadkaifshaik/dzeroth/internal/comments"
	"github.com/mohamadkaifshaik/dzeroth/internal/innercircle"
	"github.com/mohamadkaifshaik/dzeroth/internal/posts"
	"github.com/mohamadkaifshaik/dzeroth/internal/reactions"
	"github.com/mohamadkaifshaik/dzeroth/internal/users"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(databasePool *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux()

	// Users
	userRepository := users.NewRepository(databasePool)
	userHandler := users.NewHandler(userRepository)

	mux.HandleFunc(
		"/api/v1/me",
		userHandler.GetMe,
	)

	// Inner Circle
	innerCircleRepository := innercircle.NewRepository(databasePool)
	innerCircleHandler := innercircle.NewHandler(innerCircleRepository)

	mux.HandleFunc(
		"/api/v1/inner-circle",
		innerCircleHandler.GetMembers,
	)

	mux.HandleFunc(
		"/api/v1/inner-circle/",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				innerCircleHandler.AddMember(w, r)

			case http.MethodDelete:
				innerCircleHandler.RemoveMember(w, r)

			default:
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
			}
		},
	)

	// Posts
	postRepository := posts.NewRepository(databasePool)
	postHandler := posts.NewHandler(postRepository)

	mux.HandleFunc(
		"/api/v1/posts",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				postHandler.CreatePost(w, r)
				return
			}

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		},
	)

	// Inner Circle feed
	mux.HandleFunc(
		"/api/v1/feeds/inner-circle",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				postHandler.GetInnerCircleFeed(w, r)
				return
			}

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		},
	)

	// Comments
	commentRepository := comments.NewRepository(databasePool)
	commentHandler := comments.NewHandler(commentRepository)

	// Reactions
	reactionRepository := reactions.NewRepository(databasePool)
	reactionHandler := reactions.NewHandler(reactionRepository)

	mux.HandleFunc(
		"/api/v1/posts/",
		func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			switch {
			case strings.HasSuffix(path, "/comments"):
				switch r.Method {
				case http.MethodPost:
					commentHandler.CreateComment(w, r)

				case http.MethodGet:
					commentHandler.GetComments(w, r)

				default:
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
				}

			case strings.HasSuffix(path, "/reactions"):
				if r.Method != http.MethodPost {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				reactionHandler.AddReaction(w, r)

			case strings.HasSuffix(path, "/reactions/me"):
				if r.Method != http.MethodGet {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				reactionHandler.GetMyReactions(w, r)

			case strings.Contains(path, "/reactions/"):
				if r.Method != http.MethodDelete {
					http.Error(
						w,
						"method not allowed",
						http.StatusMethodNotAllowed,
					)
					return
				}

				reactionHandler.RemoveReaction(w, r)

			default:
				http.NotFound(w, r)
			}
		},
	)

	return mux
}
