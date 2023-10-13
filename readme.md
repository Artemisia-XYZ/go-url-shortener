# URL Shortener API with Go Fiber

## Tech Stack
- Golang
- Redis
- MySQL


## Usage
1. Start Container

```
docker compose up -d --build
```

## Endpoint

|Method |Endpoint       |Rate Limit         |Description            |
|---    |---            |---                |---                    |
|GET    |/<slash_code> |1,000 per 1 hour   |Redirect to destination|
|POST   |/api/links     |150 per 1 hour     |Create Short Link      |

## Example

### Request
|Parameter  |Type   |Description    |
|---        |---    |---            |
|slash_code	|String |(Optional) Shorten Code|
|destination|String |Redirect URL|

### Response
|Parameter  |Type   |Description    |
|---        |---    |---            |
|id	        |String	|UUIDv4         |
|slash_code	|String	|Shorten Code|
|origin	    |String	|Shortened URL|
|destination|String	|Redirect URL|
|visitors	|Integer|Clicks|
|created_at	|String	|Created time|
|updated_at	|String	|Updated time|

Example Request
```
{
    "slash_code": "test",
    "destination": "https://docs.gofiber.io/"
}
```

Example Response
```
{
    "id": "c6633797-5031-4326-9997-9a190a771399",
    "slash_code": "test",
    "origin": "http://127.0.0.1:5000/test",
    "destination": "https://docs.gofiber.io/",
    "visitors": 0,
    "created_at": "2023-10-10T12:34:56.789+07:00",
    "updated_at": "2023-10-10T12:34:56.789+07:00",
}
```