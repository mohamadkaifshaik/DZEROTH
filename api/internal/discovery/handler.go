package discovery

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	repository *Repository
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) GetFeed(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID := r.Header.Get("X-User-ID")

	if userID == "" {
		http.Error(
			w,
			"X-User-ID header is required",
			http.StatusUnauthorized,
		)
		return
	}

	posts, err := handler.repository.GetFeed(
		r.Context(),
		userID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to get discovery feed",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":        posts,
		"count":        len(posts),
		"has_more":     false,
		"is_caught_up": true,
	})
}
