package main

import (
	"fmt"
	"http-server/pkg/http1"
	"io"
)

func main() {
	fmt.Println("[INFO] Starting HTTP1.1 Client...")

	// Create a new HTTP client
	client := http1.NewClient()

	// Send a GET request to the server
	resp, err := client.Get("http://127.0.0.1:80/home")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response status code
	fmt.Println("Response")
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Status Code:", resp.StatusCode)
	fmt.Println("Response Headers:", resp.Header)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("[ERROR] ", err)
		return
	}
	fmt.Printf("Response Body:\n%s", string(body))

}
