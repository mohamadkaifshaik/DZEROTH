package provisioning

import (
	"net/http"

	"github.com/mohamadkaifshaik/dzeroth/internal/auth"
	"github.com/mohamadkaifshaik/dzeroth/internal/users"
)

type Middleware struct {
	repository *users.Repository
}

func NewMiddleware(
	repository *users.Repository,
) *Middleware {
	return &Middleware{
		repository: repository,
	}
}

func (middleware *Middleware) EnsureUser(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
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
			email, ok := auth.Email(
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

			err := middleware.repository.EnsureUser(
				r.Context(),
				userID,
				email,
			)

			if err != nil {
				http.Error(
					w,
					"failed to provision user: "+err.Error(),
					http.StatusInternalServerError,
				)
				return
			}

			next.ServeHTTP(
				w,
				r,
			)
		},
	)
}
