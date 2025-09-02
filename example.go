package webtenderApi

import (
	"fmt"
	"log"
	// you will need this
	// webtenderApi "github.com/webtender/api-client-go"
	// optionally load .env file
	// _ "github.com/joho/godotenv/autoload"
)

func ExampleUsage() {
	// Requires environment variables for WEBTENDER_API_KEY and WEBTENDER_API_SECRET
	client := NewClientDefaultsFromEnv()

	// Example GET request to list servers (paginated)
	serverListResponse, err := client.Get("/v1/servers")
	if err != nil {
		log.Fatalf("GET request failed: %v", err)
	}
	if serverListResponse.Status != 200 {
		log.Fatalf("GET request failed: %v", serverListResponse.Error)
	}

	serverList := serverListResponse.Data.([]map[string]interface{})
	fmt.Printf("Found %d servers\n", len(serverList))
	for _, server := range serverList {
		fmt.Printf("Server: %s\n", server["id"])
	}

	// Example POST request to create a new server
	createServerResponse, err := client.Post("/v1/servers", []byte(`{"name": "test-server"}`))
	if err != nil {
		log.Fatalf("POST request failed: %v", err)
	}
	if createServerResponse.Status != 200 {
		log.Fatalf("POST request failed: %v", createServerResponse.Error)
	}
	serverId := createServerResponse.Data.(map[string]interface{})["id"]
	fmt.Printf("Server created: %s\n", serverId)

	// Example DELETE request to delete a server by ID
	deleteServerResponse, err := client.Delete(fmt.Sprintf("/v1/servers/%s", serverId))
	if err != nil {
		log.Fatalf("GET request failed: %v", err)
	}
	if deleteServerResponse.Status > 299 {
		log.Fatalf("GET request failed: %v", deleteServerResponse.Error)
	}
	fmt.Printf("Server deleted: %s\n", serverId)

}
