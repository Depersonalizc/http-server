package http1

import (
	"errors"
	"net/http"
)

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

func (rr *RequestRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Find the right handler and delegate the request to it
	rr.getHandler(r).ServeHTTP(w, r)
}

func (rr *RequestRouter) getHandler(request *http.Request) http.Handler {
	// TODO: Perform some cleaning on request.URL.Path?

	handler, ok := rr.handlers[request.URL.Path]
	if !ok {
		handler = PageNotFoundHandler
	}
	return handler
}