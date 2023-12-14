package main

import (
	"fmt"
	"http-server/pkg/http1"
	"net/http"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

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
		fmt.Errorf("[ERROR] Cannot register handler function for /home: %v\n", err)
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Errorf("[ERROR] from ListenAndServe(): %v\n", err)
	}
}
