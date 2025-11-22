package main

import (
	"fmt"
	"os"

	"DeNet/utils"

	"github.com/google/uuid"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/generate_token.go <user_id>")
		fmt.Println("Example: go run scripts/generate_token.go 550e8400-e29b-41d4-a716-446655440000")
		os.Exit(1)
	}

	userIDStr := os.Args[1]
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		fmt.Printf("Error: Invalid UUID format: %v\n", err)
		os.Exit(1)
	}

	token, err := utils.GenerateToken(userID)
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated token for user %s:\n", userID)
	fmt.Println(token)
}

