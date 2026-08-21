package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey contextKey = "user_id"
	emailKey  contextKey = "email"
)

type Middleware struct {
	jwks     keyfunc.Keyfunc
	issuer   string
	audience string
}

func NewMiddleware() (*Middleware, error) {
	supabaseURL := strings.TrimRight(
		os.Getenv("SUPABASE_URL"),
		"/",
	)

	if supabaseURL == "" {
		return nil, fmt.Errorf(
			"SUPABASE_URL is required",
		)
	}

	jwksURL :=
		supabaseURL +
			"/auth/v1/.well-known/jwks.json"

	jwks, err := keyfunc.NewDefaultCtx(
		context.Background(),
		[]string{jwksURL},
	)

	if err != nil {
		return nil, fmt.Errorf(
			"failed to load Supabase JWKS: %w",
			err,
		)
	}

	return &Middleware{
		jwks:     jwks,
		issuer:   supabaseURL + "/auth/v1",
		audience: "authenticated",
	}, nil
}

func (middleware *Middleware) RequireAuth(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("AUTH MIDDLEWARE")
			fmt.Println("Method:", r.Method)
			fmt.Println("Path:", r.URL.Path)

			fmt.Println("ALL HEADERS:")
			for key, values := range r.Header {
				fmt.Printf("%s: %v\n", key, values)
			}

			header := r.Header.Get("Authorization")

			fmt.Println(
				"Has Authorization:",
				header != "",
			)

			if header == "" {
				http.Error(
					w,
					"Authorization header is required",
					http.StatusUnauthorized,
				)
				return
			}

			parts := strings.SplitN(
				header,
				" ",
				2,
			)

			if len(parts) != 2 ||
				!strings.EqualFold(
					parts[0],
					"Bearer",
				) {

				http.Error(
					w,
					"invalid Authorization header",
					http.StatusUnauthorized,
				)
				return
			}

			tokenString := parts[1]

			token, err := jwt.Parse(
				tokenString,
				middleware.jwks.Keyfunc,
				jwt.WithIssuer(
					middleware.issuer,
				),
				jwt.WithAudience(
					middleware.audience,
				),
			)

			if err != nil ||
				!token.Valid {

				http.Error(
					w,
					"invalid or expired token",
					http.StatusUnauthorized,
				)
				return
			}

			claims, ok :=
				token.Claims.(jwt.MapClaims)

			if !ok {
				http.Error(
					w,
					"invalid token claims",
					http.StatusUnauthorized,
				)
				return
			}

			subject, err :=
				claims.GetSubject()

			if err != nil ||
				subject == "" {

				http.Error(
					w,
					"invalid token subject",
					http.StatusUnauthorized,
				)
				return
			}

			email, ok := claims["email"].(string)

			if !ok || email == "" {
				http.Error(
					w,
					"invalid token email",
					http.StatusUnauthorized,
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				userIDKey,
				subject,
			)

			ctx = context.WithValue(
				ctx,
				emailKey,
				email,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

func UserID(
	ctx context.Context,
) (string, bool) {
	userID, ok :=
		ctx.Value(userIDKey).(string)

	return userID, ok
}

func Email(
	ctx context.Context,
) (string, bool) {
	email, ok := ctx.Value(emailKey).(string)

	return email, ok
}
