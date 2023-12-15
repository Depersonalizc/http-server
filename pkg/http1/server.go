package http1

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
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

/**  Client Connection **/

type ClientConn struct {
	tcpConn net.Conn
}

/**  Server  **/

type Server struct {
	Address string
	Handler http.Handler

	listener net.Listener
	clients  map[*ClientConn]struct{}
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
		server.clients = make(map[*ClientConn]struct{})
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

		cl := &ClientConn{
			//server:  server,
			tcpConn: conn,
		}
		server.clients[cl] = struct{}{}

		go server.serveClient(cl)
	}
}

func (server *Server) Close() error {
	err := server.listener.Close()

	for client := range server.clients {
		err = errors.Join(err, client.tcpConn.Close())
	}

	return err
}

func (server *Server) closeClient(c *ClientConn) error {
	fmt.Println("Closing client connection...")
	err := c.tcpConn.Close()
	delete(server.clients, c)
	return err
}

func (server *Server) serveClient(c *ClientConn) {
	rbuf := bufio.NewReader(c.tcpConn)
	wbuf := bufio.NewWriter(c.tcpConn)

	for {
		// Read the next request
		request, err := http.ReadRequest(rbuf)

		if err != nil {
			const errorHeaders = "\r\nContent-Type: text/plain; charset=utf-8\r\nConnection: close\r\n\r\n"
			const badRequest = "400 Bad Request"
			_, _ = fmt.Fprintf(c.tcpConn, "HTTP/1.1 %s%s%s", badRequest, errorHeaders, badRequest)
			break
		}

		request.Write(os.Stdout)

		// Generate the response
		respWriter := &SimpleResponseWriter{wbuf: wbuf}
		server.Handler.ServeHTTP(respWriter, request)

		if request.Header.Get("Connection") == "close" {
			break
		}
	}

	err := server.closeClient(c)
	if err != nil {
		fmt.Printf("Failed to close client connection :%v\n", err)
	}
}
