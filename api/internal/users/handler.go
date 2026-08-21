package users

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

func (handler *Handler) GetMe(
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

	user, err := handler.repository.GetByID(
		r.Context(),
		userID,
	)

	if err != nil {
		http.Error(
			w,
			"user not found",
			http.StatusNotFound,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(user)
}
