package main

import (
	"fmt"
	"os"

	"github.com/aiops/AiOpsHub/backend/pkg/jwt"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: go run generate_jwt.go <user_id> <username> <role>")
		return
	}

	userID := os.Args[1]
	username := os.Args[2]
	role := os.Args[3]

	token, err := jwt.GenerateToken(nil, userID, username, role)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== JWT Token ===")
	fmt.Printf("UserID: %s\n", userID)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Role: %s\n", role)
	fmt.Printf("Token: %s\n", token)
}
