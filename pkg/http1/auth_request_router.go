package http1

import (
	"errors"
	"net/http"
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
	
	if authFn == nil {
		return arr.requestRouter.RegisterHandlerFn(path, func(w http.ResponseWriter, r *http.Request) {
			if authFn(r.Header.Get("username"), r.Header.Get("password")) {
				hf(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				
			}
		})
	}

	return arr.requestRouter.RegisterHandlerFn(path, hf)
}

func (arr *AuthRequestRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find the right handler and delegate the request to it
}