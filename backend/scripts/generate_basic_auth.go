package main

import (
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run generate_basic_auth.go <username> <api_token>")
		fmt.Println("Example: go run generate_basic_auth.go admin 116983c934c8a8a29405f31d6f8d95d185")
		return
	}

	username := os.Args[1]
	token := os.Args[2]

	// Generate base64 encoding
	authString := username + ":" + token
	encoded := base64.StdEncoding.EncodeToString([]byte(authString))

	fmt.Println("=== Basic Auth Token ===")
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("API Token: %s\n", token)
	fmt.Printf("Auth String: %s\n", authString)
	fmt.Printf("Base64 Encoded: %s\n", encoded)
	fmt.Println("\n请将 Base64 Encoded 值设置为 MCP Server 的 auth_token")
}
