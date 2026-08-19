package reactions

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	repository *Repository
}

type AddReactionRequest struct {
	Type string `json:"type"`
}

func NewHandler(repository *Repository) *Handler {
	return &Handler{
		repository: repository,
	}
}

func (handler *Handler) AddReaction(
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

	postID := extractPostID(r.URL.Path)

	if postID == "" {
		http.Error(
			w,
			"post ID is required",
			http.StatusBadRequest,
		)
		return
	}

	var request AddReactionRequest

	err := json.NewDecoder(r.Body).Decode(&request)

	if err != nil {
		http.Error(
			w,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	request.Type = strings.TrimSpace(
		strings.ToLower(request.Type),
	)

	err = handler.repository.Add(
		r.Context(),
		userID,
		postID,
		request.Type,
	)

	if err != nil {
		switch err.Error() {
		case "invalid reaction type":
			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)

		case "you cannot access this post":
			http.Error(
				w,
				err.Error(),
				http.StatusForbidden,
			)

		default:
			http.Error(
				w,
				"failed to add reaction",
				http.StatusInternalServerError,
			)
		}

		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "reaction added",
		"type":    request.Type,
	})
}

func (handler *Handler) RemoveReaction(
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

	postID, reactionType := extractPostAndReaction(
		r.URL.Path,
	)

	if postID == "" || reactionType == "" {
		http.Error(
			w,
			"post ID and reaction type are required",
			http.StatusBadRequest,
		)
		return
	}

	err := handler.repository.Remove(
		r.Context(),
		userID,
		postID,
		reactionType,
	)

	if err != nil {
		if err.Error() == "invalid reaction type" {
			http.Error(
				w,
				err.Error(),
				http.StatusBadRequest,
			)
			return
		}

		http.Error(
			w,
			"failed to remove reaction",
			http.StatusInternalServerError,
		)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (handler *Handler) GetMyReactions(
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

	postID := extractPostID(
		strings.TrimSuffix(
			r.URL.Path,
			"/reactions/me",
		),
	)

	reactions, err := handler.repository.GetMyReactions(
		r.Context(),
		userID,
		postID,
	)

	if err != nil {
		if err.Error() == "you cannot access this post" {
			http.Error(
				w,
				err.Error(),
				http.StatusForbidden,
			)
			return
		}

		http.Error(
			w,
			"failed to get reactions",
			http.StatusInternalServerError,
		)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"reactions": reactions,
	})
}

func extractPostID(path string) string {
	const prefix = "/api/v1/posts/"

	if !strings.HasPrefix(path, prefix) {
		return ""
	}

	postID := strings.TrimPrefix(path, prefix)

	parts := strings.Split(
		strings.Trim(postID, "/"),
		"/",
	)

	if len(parts) == 0 {
		return ""
	}

	return parts[0]
}

func extractPostAndReaction(path string) (string, string) {
	const prefix = "/api/v1/posts/"

	if !strings.HasPrefix(path, prefix) {
		return "", ""
	}

	remaining := strings.TrimPrefix(path, prefix)

	parts := strings.Split(
		strings.Trim(remaining, "/"),
		"/",
	)

	if len(parts) != 3 {
		return "", ""
	}

	if parts[1] != "reactions" {
		return "", ""
	}

	return parts[0], parts[2]
}
