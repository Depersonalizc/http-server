package http1

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

// SimpleResponseWriter implements http.ResponseWriter
type SimpleResponseWriter struct {
	wbuf        *bufio.Writer
	header      http.Header
	status      int
	contentLen  int
	wroteHeader bool
}

func (w *SimpleResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *SimpleResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		log.Println("already called WriteHeader")
		return
	}

	cl := w.header.Get("Content-Length")
	if cl == "" {
		log.Printf("failed to get key Content-Length %s\n", cl)
		return
	}

	contentLen, err := strconv.Atoi(cl)
	if err != nil {
		log.Printf("failed to parse Content-Length %s: %v\n", cl, err)
		return
	} else {
		w.contentLen = contentLen
	}

	// Write the status line
	_, err = w.wbuf.WriteString(fmt.Sprintf("HTTP/1.1 %v %v\r\n", statusCode, http.StatusText(statusCode)))
	if err != nil {
		log.Printf("failed to write status line: %v\n", err)
		return
	}

	// Write header fields
	err = w.Header().Write(w.wbuf)
	if err != nil {
		log.Printf("failed to write header fields: %v\n", err)
		return
	}

	w.status = statusCode
	w.wroteHeader = true
}

func (w *SimpleResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.header.Set("Content-Length", strconv.Itoa(len(p)))
		w.WriteHeader(http.StatusOK)
	}

	// Make sure len(p) == w.contentLen
	if len(p) != w.contentLen {
		return 0, fmt.Errorf(
			"Content-Length mismatch (header: %v, actual: %v)", w.contentLen, len(p))
	}

	fmt.Println("fwefwefwefwef")
	nn, err := w.wbuf.Write(append(p, []byte("\r\n")...))
	if err != nil {
		log.Printf("Failed to Write: %v\n", err)
		return nn, err
	}

	err = w.wbuf.Flush()
	return nn, err
}
