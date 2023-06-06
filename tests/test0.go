package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Name string `json:"name"`
}

type UserReturn struct {
	Id         string `json:"id"`
	Created_at string `json:"created_at"`
	Updated_at string `json:"updated_at"`
	Name       string `json:"name"`
	Api_key    string `json:"api_key"`
}

type NewFeed struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Feed struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string
	Url           string
	UserID        uuid.UUID
	LastFetchedAt sql.NullTime
}

type FeedFollow struct {
	ID        uuid.UUID
	FeedID    uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}

type FeedFollowReturn struct {
	Feed        Feed       `json:"feed"`
	Feed_follow FeedFollow `json:"feed_follow"`
}

const baseAddr = "http://localhost:8080"

func main() {
	httpTest(
		"GET /v1/readiness",                      // description string
		"GET",                                    // method string
		fmt.Sprintf("%s/v1/readiness", baseAddr), // url string
		nil,                                      // request body interface{}
		nil,                                      // request headers map[string]string
		&UserReturn{},                            // resp body *T
		nil,                                      // respStructToPrint interface{}
		nil,                                      // respHeadersToCheck []string
	)

	httpTest(
		"GET /v1/err (expect error return)",
		"GET",
		fmt.Sprintf("%s/v1/err", baseAddr),
		nil,
		nil,
		&UserReturn{},
		nil,
		nil,
	)

	newUser := UserReturn{}
	httpTest(
		"create a new user",
		"POST",
		fmt.Sprintf("%s/v1/users", baseAddr),
		User{Name: "Bob"},
		nil,
		&newUser,
		nil,
		nil,
	)

	retrieveUser := UserReturn{}
	httpTest(
		"get newly created user",             // description string
		"GET",                                // method string
		fmt.Sprintf("%s/v1/users", baseAddr), // url string
		nil,                                  // request body interface{}
		map[string]string{
			"Authorization": fmt.Sprintf("apiKey %s", newUser.Api_key),
		}, // request headers map[string]string
		&retrieveUser, // resp body *T
		nil,           // respStructToPrint interface{}
		nil,           // respHeadersToCheck []string
	)

	// check that the two users are the same, compare id
	if newUser.Id != retrieveUser.Id {
		fmt.Println("no match!")
		fmt.Println("newUser.Id:", newUser.Id)
		fmt.Println("retrieveUser.Id:", retrieveUser.Id)
		panic("created user id does not match the retrieved user's id who has the same apikey")
	}
	fmt.Println("users match")

	httpTest(
		"create feed 1",                      // description string
		"POST",                               // method string
		fmt.Sprintf("%s/v1/feeds", baseAddr), // url string
		NewFeed{
			Name: "WagsLane Blog",
			Url:  "https://wagslane.dev/index.xml",
		}, // request body interface{}
		map[string]string{
			"Authorization": fmt.Sprintf("apiKey %s", newUser.Api_key),
		}, // request headers map[string]string
		&FeedFollowReturn{}, // resp body *T
		&FeedFollowReturn{}, // respStructToPrint interface{}
		nil,                 // respHeadersToCheck []string
	)

	httpTest(
		"create feed 2",                      // description string
		"POST",                               // method string
		fmt.Sprintf("%s/v1/feeds", baseAddr), // url string
		NewFeed{
			Name: "Bootdev Blog",
			Url:  "https://blog.boot.dev/index.xml",
		}, // request body interface{}
		map[string]string{
			"Authorization": fmt.Sprintf("apiKey %s", newUser.Api_key),
		}, // request headers map[string]string
		&FeedFollowReturn{}, // resp body *T
		&FeedFollowReturn{}, // respStructToPrint interface{}
		nil,                 // respHeadersToCheck []string
	)

	httpTest(
		"create feed 3",                      // description string
		"POST",                               // method string
		fmt.Sprintf("%s/v1/feeds", baseAddr), // url string
		NewFeed{
			Name: "Geohot Blog",
			Url:  "https://geohot.github.io/blog/feed.xml",
		}, // request body interface{}
		map[string]string{
			"Authorization": fmt.Sprintf("apiKey %s", newUser.Api_key),
		}, // request headers map[string]string
		&FeedFollowReturn{}, // resp body *T
		&FeedFollowReturn{}, // respStructToPrint interface{}
		nil,                 // respHeadersToCheck []string
	)

	httpTest(
		"get all feeds",                      // description string
		"GET",                                // method string
		fmt.Sprintf("%s/v1/feeds", baseAddr), // url string
		NewFeed{
			Name: "Geohot Blog",
			Url:  "https://geohot.github.io/blog/feed.xml",
		}, // request body interface{}
		nil,
		&[]Feed{}, // resp body *T
		&[]Feed{}, // respStructToPrint interface{}
		nil,       // respHeadersToCheck []string
	)

	httpTest(
		"get all feed follows", // description string
		"GET",                  // method string
		fmt.Sprintf("%s/v1/feed_follows", baseAddr), // url string
		nil, // request body interface{}
		map[string]string{
			"Authorization": fmt.Sprintf("apiKey %s", newUser.Api_key),
		}, // request headers map[string]string
		&[]FeedFollow{}, // resp body *T
		&[]FeedFollow{}, // respStructToPrint interface{}
		nil,             // respHeadersToCheck []string
	)

}

func httpTest[T any](
	description string,
	method,
	url string,
	reqStruct interface{},
	reqHeaders map[string]string,
	respStruct *T,
	respStructToPrint interface{},
	respHeadersToCheck []string,
) *T {
	fmt.Printf("======== %v ========\n", description)
	defer fmt.Printf("======== END ========\n")

	var bodyReader io.Reader
	if reqStruct != nil {
		dat, err := json.Marshal(reqStruct)
		if err != nil {
			log.Printf("json.Marshal body: %v\n", err)
			return nil
		}
		bodyReader = bytes.NewReader(dat)
	}

	fmt.Printf("Sending %s request to %s\n", method, url)
	cacheBuster := rand.Int()
	req, err := http.NewRequest(method, fmt.Sprintf("%s?v=%v", url, cacheBuster), bodyReader)
	if err != nil {
		fmt.Printf("http.NewRequest: %v\n", err)
		return nil
	}

	for header, value := range reqHeaders {
		req.Header.Set(header, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("http.Do: %v\n", err)
		return nil
	}
	defer resp.Body.Close()

	fmt.Println("Response received!")
	fmt.Println("Status code:", resp.StatusCode)
	if resp.StatusCode > 299 {
		return nil
	}

	if len(respHeadersToCheck) > 0 {
		fmt.Println("Headers:")
	}
	for _, respHeaderToCheck := range respHeadersToCheck {
		fmt.Printf(" - %s: %s\n", respHeaderToCheck, resp.Header.Get(respHeaderToCheck))
	}

	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("io.ReadAll: %v\n", err)
		return nil
	}

	if respStruct == nil {
		fmt.Printf("Response body:\n%s\n", string(dat))
		return nil
	}

	err = json.Unmarshal(dat, respStruct)
	if err != nil {
		log.Printf("json.Unmarshal: %v\n", err)
		return nil
	}
	if respStructToPrint != nil {
		err = json.Unmarshal(dat, respStructToPrint)
		if err != nil {
			log.Printf("json.Unmarshal: %v\n", err)
			return nil
		}
		parsedDat, err := json.MarshalIndent(respStructToPrint, "", "  ")
		if err != nil {
			log.Printf("json.MarshalIndent: %v\n", err)
			return nil
		}
		fmt.Printf("Parsed resp body: %s\n", string(parsedDat))
	}
	return respStruct
}
