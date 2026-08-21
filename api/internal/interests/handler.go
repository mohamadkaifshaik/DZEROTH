package interests

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mohamadkaifshaik/dzeroth/internal/auth"
)

type Handler struct {
	repository *Repository
}

type interestRequest struct {
	InterestID string `json:"interest_id"`
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	interests, err := handler.repository.GetAll(
		r.Context(),
	)

	if err != nil {
		http.Error(
			w,
			"failed to get interests",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"interests": interests,
	})
}

func (handler *Handler) GetMine(
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

	interests, err := handler.repository.GetForUser(
		r.Context(),
		userID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to get user interests",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"interests": interests,
	})
}

func (handler *Handler) Add(
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

	var request interestRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	request.InterestID = strings.TrimSpace(
		request.InterestID,
	)

	if request.InterestID == "" {
		http.Error(
			w,
			"interest_id is required",
			http.StatusBadRequest,
		)
		return
	}

	err := handler.repository.AddForUser(
		r.Context(),
		userID,
		request.InterestID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to add interest",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "interest added",
	})
}

func (handler *Handler) Remove(
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

	const prefix = "/api/v1/me/interests/"

	if !strings.HasPrefix(r.URL.Path, prefix) {
		http.Error(
			w,
			"interest ID is required",
			http.StatusBadRequest,
		)
		return
	}

	interestID := strings.TrimPrefix(
		r.URL.Path,
		prefix,
	)

	interestID = strings.TrimSpace(
		interestID,
	)

	if interestID == "" {
		http.Error(
			w,
			"interest ID is required",
			http.StatusBadRequest,
		)
		return
	}

	err := handler.repository.RemoveForUser(
		r.Context(),
		userID,
		interestID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to remove interest",
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
