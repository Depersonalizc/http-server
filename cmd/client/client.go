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
	client.SetKeepAlive(false, 0 , 0)

	// Send a GET request to the server
	resp, err := client.Get("http://127.0.0.1:80/")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()
	// Print the response
	fmt.Println("Response:")
	resp.Write(os.Stdout)
	fmt.Println()

	// wait for 30 seconds
	timer := time.NewTicker(15 * time.Second)
	<-timer.C

	client.SetKeepAlive(true, 0, 0)
	// Send a GET request to the server with basic authentication
	resp, err = client.Get("http://127.0.0.1:80/admin")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	// Print the response
	fmt.Println("Response:")
	resp.Write(os.Stdout)

}
