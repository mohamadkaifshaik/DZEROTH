package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mohamadkaifshaik/dzeroth/internal/database"
	"github.com/mohamadkaifshaik/dzeroth/internal/innercircle"
	"github.com/mohamadkaifshaik/dzeroth/internal/posts"
	"github.com/mohamadkaifshaik/dzeroth/internal/users"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to the database
	databasePool, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	defer databasePool.Close()

	// Initialize the user repository and handler
	userRepository := users.NewRepository(databasePool)
	userHandler := users.NewHandler(userRepository)

	http.HandleFunc("/api/v1/me", userHandler.GetMe)

	// Initialize the inner circle repository and handler
	innerCircleRepository := innercircle.NewRepository(databasePool)
	innerCircleHandler := innercircle.NewHandler(innerCircleRepository)

	http.HandleFunc(
		"/api/v1/inner-circle",
		innerCircleHandler.GetMembers,
	)

	http.HandleFunc(
		"/api/v1/inner-circle/",
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case http.MethodPost:
				innerCircleHandler.AddMember(w, r)

			case http.MethodDelete:
				innerCircleHandler.RemoveMember(w, r)

			default:
				http.Error(
					w,
					"method not allowed",
					http.StatusMethodNotAllowed,
				)
			}
		},
	)

	// Initialize the post repository and handler
	postRepository := posts.NewRepository(databasePool)
	postHandler := posts.NewHandler(postRepository)

	http.HandleFunc(
		"/api/v1/posts",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost {
				postHandler.CreatePost(w, r)
				return
			}

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		},
	)

	http.HandleFunc(
		"/api/v1/feeds/inner-circle",
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet {
				postHandler.GetInnerCircleFeed(w, r)
				return
			}

			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
		},
	)

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		err := databasePool.Ping(r.Context())

		if err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "Dzeroth API and database are running")
	})

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	fmt.Printf("Dzeroth API running on http://localhost:%s\n", port)

	err = http.ListenAndServe(":"+port, nil)

	if err != nil {
		log.Fatal(err)
	}
}
