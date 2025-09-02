# WebTender API Client for Go

This is a Go implementation of [WebTender's REST API](https://webtender.host/api) HTTP Client.
You should use this lightweight package to easily start using the API.

You can create your API Key and secret from the [API Key Manager in the WebTender Console](https://console.webtender.host/api-keys).

## Requirements

Go is required to use this package.
Tested on Go 1.24.1

## Authentication

WebTender's API uses API keys to authenticate requests along with a [HMAC signature (see implementation code.)](https://github.com/search?q=repo%3AWebTender%2Fapi-client-go%20SignRequest&type=code) The signature may be tricky to implement, so we recommend using this package to get started.

The client will automatically look for a local `.env` file to get the API key and secret.

Place your API key in a .env
```env
WEBTENDER_API_KEY=your-api-key
WEBTENDER_API_SECRET=your-api-secret
```

Simply construct
```go
import (
	// Load .env file
	_ "github.com/joho/godotenv/autoload"
)

wtClient := webtenderApi.NewClientDefaultsFromEnv()
```

Alternatively use the constructor to pass the API key and secret.

```go
wtClient := webtenderApi.NewClient(webtenderApi.Config{
    APIKey: "your-api-key",
    APISecret: "your-api-secret",
})
```

## Make GET, POST, PATCH, PUT, DELETE requests

The client exposes the following methods to make requests to the API.

```go
wtClient.Get(path: string) (*ApiResponse, error);
wtClient.Post(path, body: []byte) (*ApiResponse, error);
wtClient.Patch(path, body: []byte) (*ApiResponse, error);
wtClient.Put(path, body: []byte) (*ApiResponse, error);
wtClient.Delete(path: string) (*ApiResponse, error);
```

## Raw Request

If you need more control over the request, you can use the raw request methods.

```go
req := wtClient.NewRequest(method, path string, body []byte) (*http.Request, error);
```

Sign an existing request with the client.
```go
wtClient.SignRequest(req: *http.Request) error;
```

## Example

See [example.go](example.go) for a full example code.

## Testing

You can override the default API endpoint by setting the `WEBTENDER_API_BASE_URL` environment variable.

```env
WEBTENDER_API_BASE_URL=https://api.webtender.host/api
```

Or in your code:

```go
wtClient := webtenderApi.NewClient(webtenderApi.Config{
	BaseURL: "https://api.webtender.host/api",
    // ...
})
```