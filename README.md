# RSS/Atom/JSON blog feed aggregator

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
- if url already exists in the db, either from one of the current user's previous feeds or in another user's feed, will just respond with 200 and `null`
- by default also creates a feedfollow that follows this feed

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