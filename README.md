# RSS/Atom blog feed aggregator

What is this?
> An rss blog feed aggregator. Users create accounts. Users authenticate. Users then can create feeds that they want to keep in touch with (give url of blog rss). Then the server will automatically fetch the feeds from the website and parse through the xml received to know what posts the blog has. The User will then be able to retrieve all the posts that they would care about from their blogs. Users choose what blogs to care about through the feed_follows. That is their option to opt-in or opt-out of blog feeds that they would like to follow. The Go Webserver communicates with a PostgresSQL database.

A plain simple, insecure and only for personal use front-end is available to use with this backend which will be accessible at `localhost:8080` if the server is up.

Notes:
- haven't tested with JSON feeds
- won't work with blogs that don't have a standard, static .xml file (looking at you websites that have a dynamically generate xml file that isn't properly named with the extension `.xml`)

# Endpoints

## `POST /v1/users` - create user
request
```json
{
    "name": "examplename"
}
```
reply
```json
{
  "id": "f46f3480-ae95-4a5d-b570-81530f513acd",
  "created_at": "2023-06-01T14:57:42Z",
  "updated_at": "2023-06-01T14:57:42Z",
  "name": "examplename",
  "api_key": "e6a5a131393e624be822e8beb36707bc0917cc98f48db6118a76912b004a6a96"
}
```

## `GET /v1/users/` - get a user, need to have user apikey in Authorization header like `Authorization: apikey <key>`
response
```json
{
  "ID": "f46f3480-ae95-4a5d-b570-81530f513acd",
  "CreatedAt": "2023-06-01T14:57:42.488944Z",
  "UpdatedAt": "2023-06-01T14:57:42.488947Z",
  "Name": "examplename",
  "ApiKey": "e6a5a131393e624be822e8beb36707bc0917cc98f48db6118a76912b004a6a96"
}
```

## `POST /v1/feeds` - create a feed, need to have user apikey in Authorization header like `Authorization: apikey <key>`
request
```json
{
  "name": "The Boot.dev Blog",
  "url": "https://blog.boot.dev/index.xml"
}
```
response
```json
{
  "feed": {
    "ID": "3a12b21b-b778-4bdf-b027-c6dda54bc550",
    "CreatedAt": "2023-06-01T17:42:26.490305Z",
    "UpdatedAt": "2023-06-01T17:42:26.490305Z",
    "Name": "The Boot.dev Blog",
    "Url": "https://blog.boot.dev/index.xml",
    "UserID": "f46f3480-ae95-4a5d-b570-81530f513acd"
  },
  "feed_follow": {
    "ID": "2816a44c-3c97-44d6-9522-545f4cc963dd",
    "FeedID": "3a12b21b-b778-4bdf-b027-c6dda54bc550",
    "UserID": "f46f3480-ae95-4a5d-b570-81530f513acd",
    "CreatedAt": "2023-06-01T17:42:26.490305Z",
    "UpdatedAt": "2023-06-01T17:42:26.490305Z"
  }
}
```
Important distinction:
- if url is unique, both the `feed` and `feed_follow` will be created, response code is `201` and you should see both the struct values be correct
- if url already exists in the db, either from one of the current user's previous feeds or in another user's feed, response code is `200` and the `feed` struct will be zero values while the `feed_follow` struct is what was the only thing created

## `GET /v1/feeds` - get all feeds
response
```json
[
  {
    "ID": "0c274fc6-531f-44b5-b16d-267d7d72ee63",
    "CreatedAt": "2023-06-01T16:03:51.277168Z",
    "UpdatedAt": "2023-06-01T16:03:51.277172Z",
    "Name": "Machine Learning Mastery Blog",
    "Url": "https://machinelearningmastery.com/feed/",
    "UserID": "f46f3480-ae95-4a5d-b570-81530f513acd"
  },
  {
    "ID": "b9b61575-b928-4d9a-bee4-8f088f380adf",
    "CreatedAt": "2023-06-01T15:43:03.862732Z",
    "UpdatedAt": "2023-06-01T15:43:03.862735Z",
    "Name": "The Boot.dev Blog",
    "Url": "https://blog.boot.dev/index.xml",
    "UserID": "f46f3480-ae95-4a5d-b570-81530f513acd"
  }
]
```

## `POST /v1/feed_follows` - create a feed_follow to a specific feed, need to have user apikey in Authorization header like `Authorization: apikey <key>`
request
```json
{
    "feed_id": "3a12b21b-b778-4bdf-b027-c6dda54bc550"
}
```
response
```json
{
    "ID": "2816a44c-3c97-44d6-9522-545f4cc963dd",
    "FeedID": "3a12b21b-b778-4bdf-b027-c6dda54bc550",
    "UserID": "f46f3480-ae95-4a5d-b570-81530f513acd",
    "CreatedAt": "2023-06-01T17:42:26.490305Z",
    "UpdatedAt": "2023-06-01T17:42:26.490305Z"
  }
```

## `DELETE /v1/feed_follows/{feedFollowID}` - delete a feed_follow by its id
- without a given ID, will return 405, or if left trailing `/` 404
- with a valid feed_follow ID returns 200 and `null` body

## `GET /v1/feed_follows` - gets all the feed_follows of a user, need to have user apikey in Authorization header like `Authorization: apikey <key>`

```json
[
  {
    "ID": "0d96dc79-2d11-44ad-8788-2df95461d632",
    "FeedID": "45cbebfd-531f-4f78-9189-c6b733aac46c",
    "UserID": "f46f3480-ae95-4a5d-b570-81530f513acd",
    "CreatedAt": "2023-06-01T17:49:47.901203Z",
    "UpdatedAt": "2023-06-01T17:49:47.901203Z"
  }
]
```

## `GET /v1/posts` - get all the posts for user, need to have user apikey in Authorization header like `Authorization: apikey <key>`
Returns a list of all the posts from blogs whose feeds this user follows. If the user doesn't follow any feed, the response will be `null`.
- Accepts an optional query parameter `limit` that modifies how many blog posts to return. The posts returned are ordered descending by their publication date, so you will see all the newest posts at the top.
 


## `GET /v1/readiness` - readiness endpoint, returns 200 if server on

## `GET /v1/err` - return error code 500 if server on


# Further Notes

What are the constructs? 

Feed
> The feeds construct exists to hold urls of blogs that have ever been seen. Users create feeds which hold unique blog urls. They automatically are "opted-in" -- a feed_follow is created for them that is linked to the newly created feed. Because we assume there to be multiple users, many users may follow some of the same blogs. Therefore, it exists to benefit the server if we do not simply create new feed constructs for the same blogs over and over again. If a blog url already exists in the server from another user and a different user wants to keep in touch with that blog, they will only have a feed_follow created for them. <br>
A feed should only be deleted if the user that it belongs to is deleted. However, if there are other users who follow that feed, what will happen to the feed? the feed has an id of the user who created it, but now that user no longer exists? does the feed get its ownership transferred?  
```go
type Feed struct {
	ID            uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Name          string
	Url           string
	UserID        uuid.UUID
	LastFetchedAt sql.NullTime
}
```
Feed Follows
> These are automatically created whenever feeds are created by Users, regardless of whether a feed is actually created in the database or not. Users can delete feed_follows in order to opt-out of following a particular blog. Feeds will be fetched if there are feed_follows that exist.
```go
type FeedFollow struct {
	ID        uuid.UUID
	FeedID    uuid.UUID
	UserID    uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
}
```
Posts
> These are the constructs that hold information about posts from blogs that Users choose to follow. They are automatically constructed whenever the server fetches feeds from followed blogs. They are retrievable at demand from Users. 
```go
type Post struct {
	ID          uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Title       string
	Url         string
	Description string
	PublishedAt time.Time
	FeedID      uuid.UUID
}
``` 

When does the server fetch feeds?
> Periodically with an arbitrary time delay. The number of feeds to be fetched can be updated. It knows what feeds to fetch by checking the feeds in the feeds table and checks their "last_updated" column. If it is null or if it is older than 1 hour from the current time, it will be retrieved again.