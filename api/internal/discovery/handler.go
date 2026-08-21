package discovery

import (
	"encoding/json"
	"net/http"

	"github.com/mohamadkaifshaik/dzeroth/internal/auth"
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
	userID, ok := auth.UserID(
		r.Context(),
	)

	if !ok {
		http.Error(
			w,
			"unauthorized",
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
