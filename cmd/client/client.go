package main

import (
	"fmt"
	"http-server/pkg/http1"
	"os"
	"time"
)

func main() {
	fmt.Println("[INFO] Starting HTTP1.1 Client...")

	// Create a new HTTP client
	client := http1.NewClient()
	client.SetKeepAlive(false, 0, 0)

	// Send a GET request to the server
	fmt.Println("[INFO] Sending Get request for /index.html")
	resp, err := client.Get("http://127.0.0.1:80/")
	if err != nil {
		fmt.Println("[ERROR] Failed to Get /index.html:", err)
		return
	}
	defer resp.Body.Close()
	// Print the response
	fmt.Println("[INFO] Response:")
	resp.Header.Write(os.Stdout)
	fmt.Println("(Content body omitted)\n")

	// Wait for a while
	timer := time.NewTicker(15 * time.Second)
	<-timer.C

	client.SetKeepAlive(true, 0, 0)
	// Send a GET request to the server with basic authentication
	fmt.Println("[INFO] Sending Get request for /admin")
	resp, err = client.Get("http://127.0.0.1:80/admin")
	if err != nil {
		fmt.Println("[ERROR] Failed to Get /admin:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response
	fmt.Println("[INFO] Response:")
	resp.Write(os.Stdout)

}
