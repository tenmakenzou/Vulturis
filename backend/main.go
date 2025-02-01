package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"github.com/xristoskrik/vulturis/internal/database"
)

func main() {

	const port = ":8080"
	//loading environment variables
	godotenv.Load()

	//loading the secret key and database url
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SECRET_KEY")
	
	//open postgres db
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()

	// Initialize a new Queries instance, which provides methods to execute
	// SQL queries and commands using the database connection pool (db).
	dbQueries := database.New(db)

	//Initialize api config
	apiCfg := ApiConfig{
		DB:        dbQueries,
		SecretKey: secret,
	}

	//router
	r := chi.NewRouter()

	//cors
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
	})
	r.Use(cors.Handler)

	//endpoints
	r.Route("/api", func(r chi.Router) {
		r.Post("/users", apiCfg.UserCreateHandler)
		r.Delete("/users", apiCfg.UserDeleteHandler)
		r.Put("/users", apiCfg.UserUpdateHandler)
		r.Get("/users", apiCfg.UserGetHandler)
		r.Post("/users/login", apiCfg.UserloginHandler)
		r.Get("/users/authenticate", apiCfg.UserAuthenticateHandler)

		r.Get("/orders", apiCfg.OrdersGetHandler)

		r.Post("/products", apiCfg.ProductCreateHandler)
		r.Get("/products", apiCfg.GetAllProductsHandler)
		r.Put("/products", apiCfg.ProductUpdateHandler)
		r.Delete("/products", apiCfg.ProductDeleteHandler)

	})

	//serve
	log.Printf("on port: %s\n", port)
	log.Fatal(http.ListenAndServe(port, r))

}
