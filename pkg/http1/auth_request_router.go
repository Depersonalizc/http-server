package http1

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"encoding/base64"

)

// RequestRouter implements a HTTP request router/multiplexer
type AuthRequestRouter struct {
	requestRouter *RequestRouter
}

func NewAuthRequestRouter() *AuthRequestRouter {
	return &AuthRequestRouter{
		requestRouter: &RequestRouter{},
	}
}

type AuthFn func(username string, password string) bool

func (arr *AuthRequestRouter) RegisterHandlerFn(path string, hf HandlerFn, authFn AuthFn) error {
	if path == "" {
		return errors.New("empty path")
	}
	if hf == nil {
		return errors.New("nil handler function")
	}
	// if the handler function need to be authenticated
	if authFn == nil {
		// register the handler function with authentication
		return arr.requestRouter.RegisterHandlerFn(path, func(w http.ResponseWriter, r *http.Request) {
			// check if the request contains basic authentication
			authValsEncoded := r.Header.Get("Authorization")
			if authValsEncoded == "" || !strings.HasPrefix(authValsEncoded, "Basic ") {
				w.Header().Set("WWW-Authenticate", "Basic")
				w.WriteHeader(http.StatusUnauthorized)
				content := []byte(fmt.Sprintf("%v %v", http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))
				w.Write([]byte(content))
				return
			}
			// decode the basic authentication
			authValsEncoded = strings.TrimPrefix(authValsEncoded, "Basic ")
			authValsEncoded = strings.TrimSpace(authValsEncoded)
			authValsDecodedBytes, err := base64.StdEncoding.DecodeString(authValsEncoded)
			if err != nil {
				// handle error
				fmt.Printf("Error decoding basic authentication: %v\n", err)
			}
			authVals := strings.Split(string(authValsDecodedBytes), ":")
			// check if the basic authentication in wrong format
			if len(authVals) != 2 {
				w.Header().Set("WWW-Authenticate", "Basic")
				w.WriteHeader(http.StatusUnauthorized)
				content := []byte(fmt.Sprintf("%v %v", http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))
				w.Write([]byte(content))
				return
			}
			// authenticate if the basic authentication contains only username and password
			if authFn(authVals[0], authVals[1]) {
				// if authenticated, call the handler function
				hf(w, r)
			} else {
				// if not authenticated, return 401
				w.Header().Set("WWW-Authenticate", "Basic")
				w.WriteHeader(http.StatusUnauthorized)
				content := []byte(fmt.Sprintf("%v %v", http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized)))
				w.Write([]byte(content))
			}
		})
	}
	// register the handler function without authentication
	return arr.requestRouter.RegisterHandlerFn(path, hf)
}

func (arr *AuthRequestRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find the right handler and delegate the request to it
	arr.requestRouter.ServeHTTP(w, r)
}