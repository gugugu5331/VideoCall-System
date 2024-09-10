package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
}

func main() {
	// 生成唯一用户名
	timestamp := time.Now().Unix()

	req := RegisterRequest{
		Username: fmt.Sprintf("testuser%d", timestamp),
		Email:    fmt.Sprintf("testuser%d@example.com", timestamp),
		Password: "password123",
		Nickname: fmt.Sprintf("Test User %d", timestamp),
		Phone:    "",
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	fmt.Printf("Sending request: %s\n", string(jsonData))

	resp, err := http.Post("http://localhost:8081/api/v1/auth/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Response Headers: %v\n", resp.Header)
	fmt.Printf("Response Body Length: %d\n", len(body))
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("✅ Registration successful!")
	} else {
		fmt.Printf("❌ Registration failed with status %d\n", resp.StatusCode)
	}
}
