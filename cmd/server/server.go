package main

import (
	"fmt"
	"http-server/pkg/http1"
	"net/http"
	"time"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Length", "8")

	_, err := w.Write([]byte("fuck you"))
	if err != nil {
		fmt.Printf("[ERROR] Failed to write content: %v\n", err)
	}
}

func main() {
	fmt.Println("[INFO] Starting HTTP1.1 Server...")

	server := &http1.Server{}

	err := http1.RegisterHandlerFn("/home", homeHandler)
	if err != nil {
		fmt.Printf("[ERROR] Cannot register handler function for /home: %v\n", err)
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("[ERROR] from ListenAndServe(): %v\n", err)
	}

	// Set a timer for 1 minute
	timer := time.NewTimer(1 * time.Minute)

	// Wait for the timer to expire
	<-timer.C

	err = server.Close()
	if err != nil {
		fmt.Printf("[ERROR] from Close(): %v\n", err)
	}
}
