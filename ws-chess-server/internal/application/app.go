package application

import (
	"context"
	"net"
	"net/http"
	"ws-chess-server/internal/config"
	"ws-chess-server/internal/delivery/http/handlers"
	"ws-chess-server/internal/delivery/http/middleware"
	"ws-chess-server/internal/delivery/http/routers"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type App struct {
	cfg        *config.AppConfig
	httpServer *http.Server
	wsListener *handlers.WebsocketListener
	logger     middleware.Logger

	clients map[uuid.UUID]*websocket.Conn
}

func NewApp(cfg *config.AppConfig, logger middleware.Logger) *App {
	return &App{
		cfg:        cfg,
		logger:     logger,
		wsListener: handlers.NewWebsocketListener(cfg, logger),
		clients:    make(map[uuid.UUID]*websocket.Conn, 10),
	}
}

func (a *App) Run(ctx context.Context, router routers.Router) {
	a.SetupRoutes(router)

	a.httpServer = &http.Server{
		Addr:           ":" + a.cfg.Port,
		Handler:        router,
		MaxHeaderBytes: 1 << 10,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	a.logger.Infof("starting a server at port :%s", a.cfg.Port)
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalf("failed to run a server: %s", err)
		}
	}()

	go a.HandleConnections()

	<-ctx.Done()
	a.logger.Info("received a signal to shutdown the server")

	if err := a.Shutdown(); err != nil {
		a.logger.Fatalf("failed to shutdown a server: %s", err)
	}
}

func (a *App) Shutdown() error {
	a.logger.Info("shutting the server down...")
	a.wsListener.Close()
	return a.httpServer.Close()
}

func (a *App) SetupRoutes(router routers.Router) {
	router.GET("/ws", a.wsListener.HandleWebsocketConnection)
}

func (a *App) HandleConnections() {
	for newClient := range a.wsListener.RegisterChan() {
		generatedUUID := uuid.New()
		a.clients[generatedUUID] = newClient

		newClient.WriteMessage(websocket.TextMessage, []byte("hello world"))
	}
}
