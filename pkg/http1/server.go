package http1

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

/**  Handlers  **/
var (
	DefaultRequestRouter = &defaultRequestRouter
	PageNotFoundHandler  = &pageNotFoundHandler

	defaultRequestRouter RequestRouter
	pageNotFoundHandler  PageNotFound
)

type HandlerFn func(w http.ResponseWriter, r *http.Request)

func (hf HandlerFn) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hf(w, r)
}

func RegisterHandlerFn(path string, hf HandlerFn) error {
	return DefaultRequestRouter.RegisterHandlerFn(path, hf)
}

// PageNotFound writes a 404 page not found response
type PageNotFound struct{}

func (PageNotFound) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintln(w, "404 page not found")
}

// RequestRouter implements a HTTP request router/multiplexer
type RequestRouter struct {
	handlers map[string]http.Handler // All registered http.Handler's
}

func (rr *RequestRouter) RegisterHandlerFn(path string, hf HandlerFn) error {
	if path == "" {
		return errors.New("empty path")
	}
	if hf == nil {
		return errors.New("nil handler function")
	}

	if rr.handlers == nil {
		rr.handlers = make(map[string]http.Handler)
	} else if _, ok := rr.handlers[path]; ok {
		return errors.New("handler for " + path + "already exists")
	}

	rr.handlers[path] = hf
	return nil
}

func (rr *RequestRouter) getHandler(request *http.Request) http.Handler {
	// TODO: Perform some cleaning on request.URL.Path?

	handler, ok := rr.handlers[request.URL.Path]
	if !ok {
		handler = PageNotFoundHandler
	}
	return handler
}

func (rr *RequestRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find the right handler and delegate the request to it
	rr.getHandler(r).ServeHTTP(w, r)
}

/**  Client  **/

type Client struct {
	//server  *Server
	tcpConn net.Conn
}

/**  Server  **/

type Server struct {
	Address string
	Handler http.Handler

	listener net.Listener
	clients  map[*Client]struct{}
}

func (server *Server) ListenAndServe() error {
	addr := server.Address
	if addr == "" {
		addr = ":http"
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	return server.Serve(ln)
}

func (server *Server) Serve(listener net.Listener) error {
	// Close old listener
	if server.listener != nil {
		err := server.listener.Close()
		if err != nil {
			return err
		}
	}

	// Replace with new listener
	server.listener = listener

	// Use default request router as handler if none provided by the user
	if server.Handler == nil {
		server.Handler = DefaultRequestRouter
	}

	if server.clients == nil {
		server.clients = make(map[*Client]struct{})
	}

	// Accept-and-serve loop
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			// https://github.com/golang/go/blob/ab9d31da9e088a271e656120a3d99cd3b1103ab6/src/net/http/server.go#L3047-L3059
			var ne net.Error
			if errors.As(err, &ne) && ne.Temporary() {
				log.Printf("Cannot accept connection: %v", ne)
				continue
			}
			return err
		}

		cl := &Client{
			//server:  server,
			tcpConn: conn,
		}
		server.clients[cl] = struct{}{}

		go server.serveClient(cl)
	}
}

func (server *Server) serveClient(c *Client) {
	rbuf := bufio.NewReader(c.tcpConn)

	for {
		// Read the next request
		request, err := http.ReadRequest(rbuf)

		if err != nil {
			const errorHeaders = "\r\nContent-Type: text/plain; charset=utf-8\r\nConnection: close\r\n\r\n"
			const badRequest = "400 Bad Request"
			_, _ = fmt.Fprintf(c.tcpConn, "HTTP/1.1 %s%s%s", badRequest, errorHeaders, badRequest)
			return
		}

		// Generate the response
		respWriter := &SimpleResponseWriter{wbuf: c.tcpConn}
		server.Handler.ServeHTTP(respWriter, request)
	}
}

// SimpleResponseWriter implements http.ResponseWriter
type SimpleResponseWriter struct {
	wbuf        io.Writer
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
		log.Fatalf("failed to get key Content-Length %s\n", cl)
	}

	contentLen, err := strconv.Atoi(cl)
	if err != nil {
		log.Fatalf("failed to parse Content-Length %s: %v\n", cl, err)
	} else {
		w.contentLen = contentLen
	}

	// Write the status line
	_, err = fmt.Fprintf(w.wbuf, "HTTP/1.1 %v %v\r\n", statusCode, http.StatusText(statusCode))
	if err != nil {
		log.Fatalf("failed to write status line: %v\n", err)
	}

	// Write header fields
	if err = w.Header().Write(w.wbuf); err != nil {
		log.Fatalf("failed to write header: %v\n", err)
	}

	w.status = statusCode
	w.wroteHeader = true
}

func (w *SimpleResponseWriter) Write(p []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
		w.contentLen = len(p)
		w.header.Set("Content-Length", strconv.Itoa(w.contentLen))
	}

	// Make sure len(p) == w.contentLen
	if len(p) != w.contentLen {
		return 0, errors.New(fmt.Sprintf(
			"Content-Length mismatch (header: %v, actual %v)", w.contentLen, len(p)))
	}

	return w.wbuf.Write(p)
}
