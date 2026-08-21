package users

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohamadkaifshaik/dzeroth/internal/auth"
)

type Handler struct {
	repository *Repository
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
	AvatarURL   string `json:"avatar_url"`
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) UpdateMe(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := auth.UserID(r.Context())

	if !ok {
		http.Error(
			w,
			"unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	var request UpdateProfileRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	request.DisplayName = strings.TrimSpace(
		request.DisplayName,
	)

	request.Bio = strings.TrimSpace(
		request.Bio,
	)

	request.AvatarURL = strings.TrimSpace(
		request.AvatarURL,
	)

	if request.DisplayName == "" {
		http.Error(
			w,
			"display name is required",
			http.StatusBadRequest,
		)
		return
	}

	user, err := handler.repository.UpdateProfile(
		r.Context(),
		userID,
		request.DisplayName,
		request.Bio,
		request.AvatarURL,
	)

	if err != nil {
		http.Error(
			w,
			"failed to update profile",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(user)
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
