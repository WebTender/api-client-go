package webtenderApi

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Client represents an HTTP client with API key authentication and HMAC signing
type Client struct {
	httpClient *http.Client
	apiKey     string
	apiSecret  string
	baseURL    string
}

// Config holds the configuration for the API client
type Config struct {
	APIKey    string
	APISecret string
	BaseURL   string
	Timeout   time.Duration
}

type ApiResponse struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
	Error  error       `json:"error"`
}

// NewClient creates a new API client with the provided configuration
func NewClient(config Config) *Client {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.BaseURL == "" {
		config.BaseURL = env("WEBTENDER_API_BASE_URL", "https://api.webtender.host/api", true)
	}
	if config.APIKey == "" {
		config.APIKey = env("WEBTENDER_API_KEY", "", true)
	}
	if config.APISecret == "" {
		config.APISecret = env("WEBTENDER_API_SECRET", "", true)
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		apiKey:    config.APIKey,
		apiSecret: config.APISecret,
		baseURL:   config.BaseURL,
	}
}

func NewClientDefaultsFromEnv() *Client {
	return NewClient(Config{})
}

func env(name string, fallback string, required bool) string {
	value := os.Getenv(name)
	if value == "" {
		value = fallback
	}
	if required && value == "" {
		panic(fmt.Sprintf("%s is required", name))
	}
	return value
}

func (c *Client) generateHMACSignature(method, fullUrl string, body []byte, timestamp string) string {
	message := method + ":" + fullUrl
	bodyString := string(body)
	if bodyString == "" {
		message += ":" + timestamp
	} else {
		message += ":" + bodyString + ":" + timestamp
	}

	h := hmac.New(sha256.New, []byte(c.apiSecret))
	h.Write([]byte(message))

	return hex.EncodeToString(h.Sum(nil))
}

// Request makes an authenticated HTTP request
func (c *Client) NewRequest(method, path string, body []byte) (*http.Request, error) {
	// Construct full URL
	url := joinPaths(c.baseURL, path)

	// Create request
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	if err := c.SignRequest(req); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	return req, nil
}

func (c *Client) SignRequest(req *http.Request) error {
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	var bodyBytes []byte
	if req.Body != nil {
		var err error
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}
		req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	signature := c.generateHMACSignature(req.Method, req.URL.String(), bodyBytes, timestamp)

	req.Header.Set("X-API-Key", c.apiKey)
	req.Header.Set("X-Timestamp", timestamp)
	req.Header.Set("X-Signature", signature)

	return nil
}

func (c *Client) GetRequest(path string) (*http.Request, error) {
	return c.NewRequest("GET", path, nil)
}

func (c *Client) PostRequest(path string, body []byte) (*http.Request, error) {
	return c.NewRequest("POST", path, body)
}

func (c *Client) PatchRequest(path string, body []byte) (*http.Request, error) {
	return c.NewRequest("PATCH", path, body)
}

func (c *Client) PutRequest(path string, body []byte) (*http.Request, error) {
	return c.NewRequest("PUT", path, body)
}

func (c *Client) DeleteRequest(path string) (*http.Request, error) {
	return c.NewRequest("DELETE", path, nil)
}

// Get makes a GET request to the API and returns the status code, data, and error
func (c *Client) Get(path string) (*ApiResponse, error) {
	req, err := c.GetRequest(path)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func (c *Client) Post(path string, body []byte) (*ApiResponse, error) {
	req, err := c.PostRequest(path, body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func (c *Client) Patch(path string, body []byte) (*ApiResponse, error) {
	req, err := c.PatchRequest(path, body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func (c *Client) Put(path string, body []byte) (*ApiResponse, error) {
	req, err := c.PutRequest(path, body)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func (c *Client) Delete(path string) (*ApiResponse, error) {
	req, err := c.DeleteRequest(path)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func (c *Client) doRequest(req *http.Request) (*ApiResponse, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	apiResponse := ApiResponse{
		Status: resp.StatusCode,
		Data:   map[string]interface{}{},
		Error:  nil,
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		apiResponse.Error = err
		return &apiResponse, err
	}
	err = json.Unmarshal(body, &apiResponse.Data)
	if err != nil {
		apiResponse.Error = err
		return &apiResponse, err
	}

	if apiResponse.Status > 299 {
		apiResponse.Error = fmt.Errorf("status: %d", apiResponse.Status)
		dataMap, ok := apiResponse.Data.(map[string]interface{})
		if ok && dataMap["message"] != nil {
			apiResponse.Error = fmt.Errorf("status: %d: %s", apiResponse.Status, dataMap["message"].(string))
		}
		return &apiResponse, apiResponse.Error
	}

	return &apiResponse, nil
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func joinPaths(base, path string) string {
	return strings.TrimSuffix(base, "/") + "/" + strings.TrimPrefix(path, "/")
}
