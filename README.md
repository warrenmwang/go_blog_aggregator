# RSS blog feed aggregator


`POST /v1/users` - create user
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


`POST /v1/feeds` - create a feed, need to have user apikey in Authorization header like 
`Authorization: apikey <key>`
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
  "ID": "b9b61575-b928-4d9a-bee4-8f088f380adf",
  "CreatedAt": "2023-06-01T15:43:03.862732Z",
  "UpdatedAt": "2023-06-01T15:43:03.862735Z",
  "Name": "The Boot.dev Blog",
  "Url": "https://blog.boot.dev/index.xml",
  "UserID": "f46f3480-ae95-4a5d-b570-81530f513acd"
}
```
- if url already exists in the db, either from one of the current user's previous feeds or in another user's feed, will just respond with 200 and `null`


`GET /v1/readiness` - readiness endpoint, returns 200 if server on

`GET /v1/err` - return error code 500 if server on