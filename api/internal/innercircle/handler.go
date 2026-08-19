package innercircle

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	repository *Repository
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) GetMembers(
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

	members, err := handler.repository.GetMembers(
		r.Context(),
		userID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to get inner circle",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"members": members,
		"count":   len(members),
		"limit":   25,
	})
}

func (handler *Handler) AddMember(
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

	memberID := strings.TrimPrefix(
		r.URL.Path,
		"/api/v1/inner-circle/",
	)

	if memberID == "" || memberID == r.URL.Path {
		http.Error(
			w,
			"user ID is required",
			http.StatusBadRequest,
		)
		return
	}

	err := handler.repository.AddMember(
		r.Context(),
		userID,
		memberID,
	)

	if err != nil {
		switch err.Error() {
		case "you cannot add yourself":
			http.Error(w, err.Error(), http.StatusBadRequest)

		case "user not found":
			http.Error(w, err.Error(), http.StatusNotFound)

		case "inner circle can contain at most 25 people":
			http.Error(w, err.Error(), http.StatusConflict)

		default:
			http.Error(
				w,
				"failed to add member",
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]string{
		"message": "member added to inner circle",
	})
}

func (handler *Handler) RemoveMember(
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

	memberID := strings.TrimPrefix(
		r.URL.Path,
		"/api/v1/inner-circle/",
	)

	if memberID == "" || memberID == r.URL.Path {
		http.Error(w, "user ID is required", http.StatusBadRequest)
		return
	}

	err := handler.repository.RemoveMember(
		r.Context(),
		userID,
		memberID,
	)

	if err != nil {
		if err.Error() == "member not found in inner circle" {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(
			w,
			"failed to remove member",
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
