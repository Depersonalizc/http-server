package http1

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

type Client struct {
	KeepAlive        bool
	keepAliveTimeout int
	keepAliveMax     int
	servers          map[string]*ServerConn
	credentials      map[string]string
}

type ServerConn struct {
	conn net.Conn
}

func NewClient() *Client {
	return &Client{
		KeepAlive: true,
		servers:   make(map[string]*ServerConn),
		credentials: make(map[string]string),
	}
}

func (c *Client) SetKeepAlive(keepAlive bool, timeout int, max int) {
	if keepAlive {
		c.KeepAlive = true
		c.keepAliveTimeout = timeout
		c.keepAliveMax = max
	} else {
		c.KeepAlive = false
	}
}

func (c *Client) Close() error {
	for _, server := range c.servers {
		err := server.conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Get(urlStr string) (*http.Response, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return nil, err
	}

	return c.Do(req)
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// Get server connection
	parsedUrl, err := url.Parse(req.URL.String())
	if err != nil {
		return nil, err
	}

	// Get server connection
	if c.servers == nil {
		c.servers = make(map[string]*ServerConn)
	}
	server, ok := c.servers[parsedUrl.Host]
	if !ok {
		newConn, err := net.Dial("tcp", parsedUrl.Host)
		fmt.Println("Dialing\n", parsedUrl.Host)
		if err != nil {
			return nil, err
		}
		server = &ServerConn{conn: newConn}
		c.servers[parsedUrl.Host] = server
	}

	// Serialize request
	reqStr := c.getRequestString(req)
	fmt.Printf("Request:\n%s", reqStr)

	// Send request
	_, err = server.conn.Write([]byte(reqStr))
	if err != nil {
		return nil, err
	}

	// Read response
	resp, err := http.ReadResponse(bufio.NewReader(server.conn), req)
	if err != nil {
		return nil, err
	}
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(body))

	// Close connection if not keep-alive
	if !c.KeepAlive {
		err = server.conn.Close()
		if err != nil {
			return nil, err
		}
		delete(c.servers, parsedUrl.Host)
	}

	if resp.StatusCode == http.StatusUnauthorized && req.Header.Get("Authorization") == "" {
		return c.tryAuthenticate(req)
	}

	return resp, nil
}

func (c *Client) tryAuthenticate(req *http.Request) (*http.Response, error) {
	// Check if credentials are cached
	credential, ok := c.credentials[req.URL.String()]
	if ok {
		// Add Authorization header
		req.Header.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(credential)))
		// Retry request
		resp, err := c.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusUnauthorized {
			return resp, nil
		}
		delete(c.credentials, req.URL.String())
	}

	retry := 3
	for retry > 0 {
		// Get credentials from user
		fmt.Print("Username: ")
		var username string
		_, err := fmt.Scanln(&username)
		if err != nil {
			return nil, err
		}
		fmt.Print("Password: ")
		var password string
		_, err = fmt.Scanln(&password)
		if err != nil {
			return nil, err
		}
		// Encode credentials
		credential = fmt.Sprintf("%s:%s", username, password)
		req.Header.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(credential)))
		// Retry request
		resp, err :=  c.Do(req)
		if err != nil {
			return nil, err
		}
		// Check if authentication succeeded
		if resp.StatusCode != http.StatusUnauthorized {
			// Cache credentials and return response
			c.credentials[req.URL.String()] = credential
			return resp, nil
		} else {
			// Decrement retry count
			retry--
		}
	}
	// Return error if authentication failed
	return nil, fmt.Errorf("authentication failed")
}

func (c *Client) getRequestString(req *http.Request) string {
	req.Write(os.Stdout)

	// Build the request string
	reqStr := fmt.Sprintf("%s %s %s\r\n", req.Method, req.URL.Path, req.Proto)
	reqStr += fmt.Sprintf("Host: %s\r\n", req.Host)
	
	// Add the request headers
	for key, values := range req.Header {
		for _, value := range values {
			reqStr += fmt.Sprintf("%s: %s\r\n", key, value)
		}
	}

	// Add Connection header
	if c.KeepAlive {
		reqStr += "Connection: keep-alive\r\n"
		if c.keepAliveTimeout > 0 && c.keepAliveMax > 0 {
			reqStr += "Keep-Alive: timeout=" + strconv.Itoa(c.keepAliveTimeout) +
				", max=" + strconv.Itoa(c.keepAliveMax) + "\r\n"
		} else if c.keepAliveTimeout > 0 {
			reqStr += "Keep-Alive: timeout=" + strconv.Itoa(c.keepAliveTimeout) + "\r\n"
		} else if c.keepAliveMax > 0 {
			reqStr += "Keep-Alive: max=" + strconv.Itoa(c.keepAliveMax) + "\r\n"
		}
	} else {
		reqStr += "Connection: close\r\n"
	}

	// Add a blank line to separate headers from body
	reqStr += "\r\n"

	// Add the request body
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("Error reading request body: %v\n", err)
		}
		reqStr += string(body)
	}
	
	// Return the request string
	return reqStr
}
