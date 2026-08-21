package posts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mohamadkaifshaik/dzeroth/internal/auth"
)

type Handler struct {
	repository *Repository
}

type CreatePostRequest struct {
	Content     string   `json:"content"`
	Visibility  string   `json:"visibility"`
	InterestIDs []string `json:"interest_ids"`
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) CreatePost(
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

	var request CreatePostRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	request.Content = strings.TrimSpace(request.Content)

	post, err := handler.repository.Create(
		r.Context(),
		userID,
		request.Content,
		request.Visibility,
		request.InterestIDs,
	)

	if err != nil {
		if strings.Contains(err.Error(), "cannot be empty") ||
			strings.Contains(err.Error(), "invalid post visibility") {

			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)

			return
		}

		http.Error(
			w,
			fmt.Sprintf("failed to create post: %v", err),
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(post)
}

func (handler *Handler) GetInnerCircleFeed(
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

	feed, err := handler.repository.GetInnerCircleFeed(
		r.Context(),
		userID,
	)

	if err != nil {
		http.Error(
			w,
			"failed to get inner circle feed",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"posts":        feed,
		"count":        len(feed),
		"has_more":     false,
		"is_caught_up": true,
	})
}
