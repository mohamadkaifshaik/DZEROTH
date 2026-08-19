package comments

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	repository *Repository
}

type CreateCommentRequest struct {
	Content string `json:"content"`
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) CreateComment(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		http.Error(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	postID := strings.TrimPrefix(
		r.URL.Path,
		"/api/v1/posts/",
	)

	postID = strings.TrimSuffix(postID, "/comments")

	if postID == "" {
		http.Error(w, "post ID is required", http.StatusBadRequest)
		return
	}

	var request CreateCommentRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	request.Content = strings.TrimSpace(request.Content)

	comment, err := handler.repository.Create(
		r.Context(),
		userID,
		postID,
		request.Content,
	)

	if err != nil {
		if strings.Contains(err.Error(), "cannot access") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		if strings.Contains(err.Error(), "cannot be empty") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(
			w,
			"failed to create comment",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(comment)
}

func (handler *Handler) GetComments(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		http.Error(w, "X-User-ID header is required", http.StatusUnauthorized)
		return
	}

	postID := strings.TrimPrefix(
		r.URL.Path,
		"/api/v1/posts/",
	)

	postID = strings.TrimSuffix(postID, "/comments")

	comments, err := handler.repository.GetForPost(
		r.Context(),
		userID,
		postID,
	)

	if err != nil {
		if strings.Contains(err.Error(), "cannot access") {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		http.Error(
			w,
			"failed to get comments",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"comments": comments,
		"count":    len(comments),
	})
}
