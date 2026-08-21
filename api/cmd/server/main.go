package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/mohamadkaifshaik/dzeroth/internal/database"
	httpapi "github.com/mohamadkaifshaik/dzeroth/internal/http"
)

func main() {
	err := godotenv.Load()

	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	databasePool, err := database.Connect()

	if err != nil {
		log.Fatal(err)
	}

	defer databasePool.Close()

	fmt.Println(">>> RUNNING UPDATED DZEROTH API <<<")

	router := httpapi.NewRouter(databasePool)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	routerWithHealth := http.NewServeMux()

	routerWithHealth.HandleFunc(
		"/health",
		func(w http.ResponseWriter, r *http.Request) {
			err := databasePool.Ping(r.Context())

			if err != nil {
				http.Error(
					w,
					"database unavailable",
					http.StatusServiceUnavailable,
				)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(
				w,
				"Dzeroth API and database are running",
			)
		},
	)

	routerWithHealth.Handle(
		"/",
		router,
	)

	fmt.Printf(
		"Dzeroth API running on http://localhost:%s\n",
		port,
	)

	err = http.ListenAndServe(
		":"+port,
		routerWithHealth,
	)

	if err != nil {
		log.Fatal(err)
	}
}
