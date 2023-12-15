package main

import (
	"fmt"
	"http-server/pkg/http1"
	"net/http"
	"os"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	content, err := os.ReadFile("resource/index.html")
	if err != nil {
		fmt.Printf("[ERROR] Failed to read file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		content = []byte(fmt.Sprintf("%v %v", http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)))
		w.Write([]byte(content))
	} else {
		w.Write(content)
	}
}

func resourceMemeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "image/jpeg")

	content, err := os.ReadFile("resource/meme.jpg")
	if err != nil {
		fmt.Printf("[ERROR] Failed to read file: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		content = []byte(fmt.Sprintf("%v %v", http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError)))
		w.Write(content)
	} else {
		w.Write(content)
	}
}

func main() {
	fmt.Println("[INFO] Starting HTTP1.1 Server...")

	server := &http1.Server{}

	err := http1.RegisterHandlerFn("/index.html", indexHandler)
	if err != nil {
		fmt.Printf("[ERROR] Cannot register handler function for /index.html: %v\n", err)
	}

	err = http1.RegisterHandlerFn("/", indexHandler)
	if err != nil {
		fmt.Printf("[ERROR] Cannot register handler function for /index.html: %v\n", err)
	}

	err = http1.RegisterHandlerFn("/resource/meme.jpg", resourceMemeHandler)
	if err != nil {
		fmt.Printf("[ERROR] Cannot register handler function for /resource/meme.jpg: %v\n", err)
	}

	err = server.ListenAndServe()
	if err != nil {
		fmt.Printf("[ERROR] from ListenAndServe(): %v\n", err)
	}

	err = server.Close()
	if err != nil {
		fmt.Printf("[ERROR] from Close(): %v\n", err)
	}
}
