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
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/mmcdole/gofeed"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

type errorBody struct {
	Error string `json:"error"`
}

type apiConfig struct {
	DB           *database.Queries
	FetchedFeeds []interface{}
}

// wrapper for respondWithJSON for sending errors as the interface used to be converted to json
func respondWithError(w http.ResponseWriter, code int, err error) {
	respondWithJSON(w, code, errorBody{Error: err.Error()})
	log.Println("responded with err:", err.Error())
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

// get JWT/APIKEY from the "Authorization" header
// expects format - Authorization: Bearer <token> / Authorization: ApiKey <key>
// where "Authorization" is the header name
func getAuthTokenFromHeader(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	splitAuthHeader := strings.Split(authHeader, " ")
	if len(splitAuthHeader) != 2 {
		return "", errors.New("no token provided")
	}
	tokenString := splitAuthHeader[1]
	return tokenString, nil
}

type authedHandler func(http.ResponseWriter, *http.Request, database.User)

// gets a user only if authenticated, that is there is a valid apikey for the user present
// in the Authorization header
// returns that authenticated user
func (cfg *apiConfig) middlewareAuth(handler authedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// get apikey from header
		apikey, err := getAuthTokenFromHeader(r)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			log.Println(err)
			return
		}

		// get user
		user, err := cfg.DB.GetUser(context.Background(), apikey)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			log.Println(err)
			return
		}

		handler(w, r, user)
	}
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

	// make sure name isn't empty
	if len(params.Name) == 0 {
		respondWithError(w, http.StatusUnauthorized, errors.New("name cannot be empty"))
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
		Api_key:    databaseUser.ApiKey,
	}

	// respond with the newly created user's info
	respondWithJSON(w, http.StatusOK, retVal)
}

// GET /v1/users/
// needs Authorization: ApiKey <key>
// get a user by their apikey
func (apiCfg apiConfig) getUserHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, http.StatusOK, user)
}

// POST /v1/feeds
// needs Authorization: ApiKey <key>
// create a new feed and by default a new feed_follow to this feed as well
// even is url is duplicate, should create a new feed_follow
func (apiCfg apiConfig) createFeedHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}
	type returnVal struct {
		Feed        database.Feed       `json:"feed"`
		Feed_follow database.FeedFollow `json:"feed_follow"`
	}

	// decode the user from JSON into go struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("decoding json went wrong"))
		return
	}

	// make sure name and url are not empty
	if len(params.Name) == 0 || len(params.Url) == 0 {
		respondWithError(w, http.StatusUnauthorized, errors.New("name and url cannot be empty"))
		return
	}

	// generate new feed's uuid
	newFeedUUID, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Error generating UUID: %v\n", err)
		return
	}

	currTime := time.Now()

	// create the new Feed
	newFeed := database.CreateFeedParams{
		ID:        newFeedUUID,
		CreatedAt: currTime,
		UpdatedAt: currTime,
		Name:      params.Name,
		Url:       params.Url,
		UserID:    user.ID,
	}

	// put the feed in the db
	createdFeed, err := apiCfg.DB.CreateFeed(context.Background(), newFeed)
	if err != nil {
		// if err is a duplicate url, create a new feed follow
		// create a new feed_follow that has the feedid of the existing feed with the same id in the db
		if err.Error() == "pq: duplicate key value violates unique constraint \"feeds_url_key\"" {
			log.Println("duplicate url, create new feed_follow to existing feed, respond with just the feed_follow")

			// find the feed that already exists whose url is the one that we have
			allFeeds, err := apiCfg.DB.GetFeeds(context.Background())
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, err)
				return
			}

			var existingFeedToFind database.Feed
			for _, feed := range allFeeds {
				if newFeed.Url == feed.Url {
					existingFeedToFind = feed
				}
			}

			newFeedFollowUUID, err := uuid.NewRandom()
			if err != nil {
				log.Fatalf("Error generating UUID: %v\n", err)
				return
			}
			newFeedFollow := database.CreateFeedFollowParams{
				ID:        newFeedFollowUUID,
				FeedID:    existingFeedToFind.ID,
				UserID:    user.ID,
				CreatedAt: currTime,
				UpdatedAt: currTime,
			}

			// save new feed_follow to db
			createdFeedFollow, err := apiCfg.DB.CreateFeedFollow(context.Background(), newFeedFollow)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, err)
				return
			}
			respondWithJSON(w, http.StatusOK, createdFeedFollow)
		} else {
			// an actual error
			respondWithError(w, http.StatusInternalServerError, err)
		}
	} else {
		// otherwise, this is a unique, new feed (url), next create a new feed_follow
		// generate new feed_follow's uuid
		newFeedFollowUUID, err := uuid.NewRandom()
		if err != nil {
			log.Fatalf("Error generating UUID: %v\n", err)
			return
		}

		// create the new feed follow
		newFeedFollow := database.CreateFeedFollowParams{
			ID:        newFeedFollowUUID,
			FeedID:    newFeedUUID,
			UserID:    user.ID,
			CreatedAt: currTime,
			UpdatedAt: currTime,
		}

		// save new feed follow to db
		createdFeedFollow, err := apiCfg.DB.CreateFeedFollow(context.Background(), newFeedFollow)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		// respond with acknowledgement that we created both a new feed and a new feed follow
		respondWithJSON(w, http.StatusOK, returnVal{
			Feed:        createdFeed,
			Feed_follow: createdFeedFollow,
		})
	}
}

// GET /v1/feeds
// retrieve all feeds, don't need to be authed
func (apiCfg apiConfig) getAllFeedsHandler(w http.ResponseWriter, r *http.Request) {
	allFeeds, err := apiCfg.DB.GetFeeds(context.Background())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, allFeeds)
}

// POST /v1/feed_follows
// authed
// expects a feed_id
func (apiCfg apiConfig) createFeedFollowHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Feed_id string `json:"feed_id"`
	}

	// decode the user from JSON into go struct
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, errors.New("decoding json went wrong"))
		return
	}

	// make sure feed_id is present
	if len(params.Feed_id) == 0 {
		respondWithError(w, http.StatusUnauthorized, errors.New("feed_id cannot be empty"))
		return
	}

	// generate new feed_follow's uuid
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Error generating UUID: %v\n", err)
		return
	}

	// get feed_id
	parsedFeedId, err := uuid.Parse(params.Feed_id)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	// create new feedfollow
	currTime := time.Now()
	newFeedFollow := database.CreateFeedFollowParams{
		ID:        newUUID,
		FeedID:    parsedFeedId,
		UserID:    user.ID,
		CreatedAt: currTime,
		UpdatedAt: currTime,
	}

	// store in db
	createdFeedFollow, err := apiCfg.DB.CreateFeedFollow(context.Background(), newFeedFollow)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}

	respondWithJSON(w, http.StatusOK, createdFeedFollow)
}

// GET /v1/feed_follows
// authed
// get all feed followers from the authed user
func (apiCfg apiConfig) getFeedFollowsHandler(w http.ResponseWriter, r *http.Request, user database.User) {
	feedFollowers, err := apiCfg.DB.GetFeedFollows(context.Background(), user.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, feedFollowers)
}

// DELETE /v1/feed_follows/{feedFollowID}
// deletes the single feed follower specified by its id
func (apiCfg apiConfig) deleteFeedFollowHandler(w http.ResponseWriter, r *http.Request) {
	idFeedToDelete := chi.URLParam(r, "feedFollowID")
	parsedFeedId, err := uuid.Parse(idFeedToDelete)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	err = apiCfg.DB.DeleteFeedFollow(context.Background(), parsedFeedId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err)
		return
	}
	respondWithJSON(w, http.StatusOK, nil)
}

// download the .xml file from the url
// there exists 3 possible formats: RSS, Atom, JSON feed
// for now, just do RSS
func getRSSFromURL(url string) (interface{}, error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(url)
	if err != nil {
		return nil, err
	}
	return feed, nil
}

// TODO: write the test for this, manual testing says seems like it works
// continuously pull things from the feed urls
// delay is in seconds
func (apiCfg apiConfig) feedFetcherWorker(delay int, fetchBatchSize int32) {
	go func(delay int, fetchBatchSize int32) {
		for {
			log.Println("fetching new feeds...")
			// get all feeds to be fetched
			feedsToUpdate, err := apiCfg.DB.GetNextFeedsToFetch(context.Background(), fetchBatchSize)
			if err != nil {
				log.Println("feedFetcherWorker: ", err)
				return
			}

			log.Printf("fetching %d feeds this round...\n", len(feedsToUpdate))

			for _, feed := range feedsToUpdate {
				// update the db that the feeds were got (updated_at, last_fetched_at)
				currTime := time.Now()
				apiCfg.DB.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
					ID: feed.ID,
					LastFetchedAt: sql.NullTime{
						Time:  currTime,
						Valid: true,
					},
					UpdatedAt: currTime,
				})

				// get rss
				data, err := getRSSFromURL(feed.Url)
				if err != nil {
					log.Println("feedFetcherWorker: ", err)
					return
				}

				// store them
				apiCfg.FetchedFeeds = append(apiCfg.FetchedFeeds, data)
			}

			// log.Println(apiCfg.FetchedFeeds)

			time.Sleep(time.Duration(delay) * time.Second)
		}
	}(delay, fetchBatchSize)
}

// get posts from user
// func (apiCfg apiConfig)

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

	v1Router.Post("/users", apiCfg.createUserHandler)                    // create a new user
	v1Router.Get("/users", apiCfg.middlewareAuth(apiCfg.getUserHandler)) // get a user using apikey

	v1Router.Post("/feeds", apiCfg.middlewareAuth(apiCfg.createFeedHandler)) // create a new feed for the authed user
	v1Router.Get("/feeds", apiCfg.getAllFeedsHandler)                        // get all feeds

	v1Router.Post("/feed_follows", apiCfg.middlewareAuth(apiCfg.createFeedFollowHandler)) // create a new feed follow for the authed user
	v1Router.Get("/feed_follows", apiCfg.middlewareAuth(apiCfg.getFeedFollowsHandler))    // get all the feed follows for the authed user
	v1Router.Delete("/feed_follows/{feedFollowID}", apiCfg.deleteFeedFollowHandler)       // delete a feed follow

	// continuously fetch feeds
	apiCfg.feedFetcherWorker(10, 10)

	log.Println("launching server")
	srv := http.Server{
		Addr:    fmt.Sprintf(":%v", port),
		Handler: router,
	}
	srv.ListenAndServe()
}
