package routers

import (
	"net/http"
	"ws-chess-server/internal/delivery/http/middleware"
	"ws-chess-server/internal/delivery/http/response"
	"ws-chess-server/pkg/logger"
)

type Handler = func(w http.ResponseWriter, req *http.Request) error

type Router interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
	GET(relativePath string, handlers ...Handler)
	POST(relativePath string, handlers ...Handler)
	DELETE(relativePath string, handlers ...Handler)
	PATCH(relativePath string, handlers ...Handler)
	PUT(relativePath string, handlers ...Handler)
}

type DefaultRouter struct {
	http.Handler

	logger   logger.Logger
	listener *http.ServeMux
}

func NewDefaultRouter(logger logger.Logger) *DefaultRouter {
	mux := http.NewServeMux()

	return &DefaultRouter{
		listener: mux,
		logger:   logger,
	}
}

func (r *DefaultRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.listener.ServeHTTP(w, req)
}

func (r *DefaultRouter) GET(relativePath string, handlers ...Handler) {
	r.listener.HandleFunc(relativePath, r.makeMiddlewareChain(handlers...))
}

func (r *DefaultRouter) POST(relativePath string, handlers ...Handler) {
	r.listener.HandleFunc(relativePath, r.makeMiddlewareChain(handlers...))
}

func (r *DefaultRouter) DELETE(relativePath string, handlers ...Handler) {
	r.listener.HandleFunc(relativePath, r.makeMiddlewareChain(handlers...))
}

func (r *DefaultRouter) PATCH(relativePath string, handlers ...Handler) {
	r.listener.HandleFunc(relativePath, r.makeMiddlewareChain(handlers...))
}

func (r *DefaultRouter) PUT(relativePath string, handlers ...Handler) {
	r.listener.HandleFunc(relativePath, r.makeMiddlewareChain(handlers...))
}

func (r *DefaultRouter) makeMiddlewareChain(handlers ...Handler) http.HandlerFunc {
	handler := func(w http.ResponseWriter, r *http.Request) {
		for i := range handlers {
			if err := handlers[i](w, r); err != nil {
				handleError(w, err)
				return
			}
		}
	}

	return middleware.LoggerMiddleware(r.logger, middleware.PanicRecovery(handler))
}

func handleError(w http.ResponseWriter, err error) {
	switch err.(type) {
	default:
		response.Error(w, err, http.StatusInternalServerError)
	}
}
