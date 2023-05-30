package main

import (
	"blog_aggregator/internal/database"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type errorBody struct {
	Error string `json:"error"`
}

type apiConfig struct {
	DB *database.Queries
}

// wrapper for respondWithJSON for sending errors as the interface used to be converted to json
func respondWithError(w http.ResponseWriter, code int, err error) {
	respondWithJSON(w, code, errorBody{Error: err.Error()})
}

// handles http requests and return json
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	response, err := json.Marshal(payload)
	if err != nil {
		log.Fatal(err)
	}
	w.Write(response)
}

// GET /v1/readiness
// returns an ok to note that the server is up and running
func readinessHandler(w http.ResponseWriter, r *http.Request) {
	type retVal struct {
		Status string `json:"status"`
	}
	respondWithJSON(w, 200, retVal{Status: "ok"})
}

// GET /v1/err
// just returns an error with code 500
func errorHandler(w http.ResponseWriter, r *http.Request) {
	respondWithError(w, 500, errors.New("Internal Server Error"))
}

// POST /v1/users
// create a new user
// create the user, store it in the db, and then return a JSON body with
// same info: id, created_at, updated_at, name
func (apiCfg apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}
	type returnValue struct {
		Id         string `json:"id"`
		Created_at string `json:"created_at"`
		Updated_at string `json:"updated_at"`
		Name       string `json:"name"`
		Api_key    string `json:"api_key"`
	}

	// decode the user from JSON into go struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("decoding json went wrong"))
		return
	}

	// generate new user's uuid
	uuid, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Error generating UUID: %v\n", err)
		return
	}

	// create the user
	newUser := database.CreateUserParams{
		ID:        uuid,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      params.Name,
	}

	// put the user in the db
	databaseUser, err := apiCfg.DB.CreateUser(context.Background(), newUser)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		log.Println(err)
		return
	}

	// create the ret val
	retVal := returnValue{
		Id:         databaseUser.ID.String(),
		Created_at: databaseUser.CreatedAt.Format(time.RFC3339),
		Updated_at: databaseUser.UpdatedAt.Format(time.RFC3339),
		Name:       databaseUser.Name,
		Api_key: databaseUser.ApiKey
	}

	// respond with the newly created user's info
	respondWithJSON(w, http.StatusOK, retVal)
}

// GET /v1/users/
// needs Authorization: ApiKey <key>
func (apiCfg apiConfig) getUserHandler(w http.ResponseWriter, r *http.Request) {

}

func main() {
	godotenv.Load() // load .env
	port := os.Getenv("PORT")
	dbURL := os.Getenv("DBURL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("couldn't connect to db, error:", err)
	}
	dbQueries := database.New(db)

	apiCfg := apiConfig{
		DB: dbQueries,
	}

	router := chi.NewRouter()
	router.Use(cors.AllowAll().Handler)

	v1Router := chi.NewRouter()
	router.Mount("/v1", v1Router)

	v1Router.Get("/readiness", readinessHandler)
	v1Router.Get("/err", errorHandler)

	v1Router.Post("/users", apiCfg.createUserHandler) // create a new user
	v1Router.Get("/users", apiCfg.getUserHandler)     // get a user using apikey

	srv := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: router,
	}
	srv.ListenAndServe()

}
