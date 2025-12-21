package application

import (
	"context"
	"net"
	"net/http"
	"ws-chess-server/internal/config"
	"ws-chess-server/internal/delivery/http/handlers"
	"ws-chess-server/internal/delivery/http/middleware"
	"ws-chess-server/internal/delivery/http/routers"
)

type App struct {
	cfg    *config.AppConfig
	server *http.Server
	logger middleware.Logger
}

func NewApp(cfg *config.AppConfig, logger middleware.Logger) *App {
	return &App{cfg: cfg, logger: logger}
}

func (a *App) Run(ctx context.Context, router routers.Router) {
	a.server = &http.Server{
		Addr:           ":" + a.cfg.Port,
		Handler:        router,
		MaxHeaderBytes: 1 << 10,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	a.SetupRoutes(router)

	a.logger.Printf("starting a server at port :%s", a.cfg.Port)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalf("failed to run a server: %s", err)
		}
	}()

	<-ctx.Done()
	a.logger.Println("received a signal to shutdown the server")

	if err := a.Shutdown(); err != nil {
		a.logger.Fatalf("failed to shutdown a server: %s", err)
	}
}

func (a *App) Shutdown() error {
	a.logger.Println("shutting the server down...")
	return a.server.Close()
}

func (a *App) SetupRoutes(router routers.Router) {
	wsListener := handlers.NewWebsocketListener(a.cfg)

	router.GET("/ws", wsListener.HandleWebsocketConnection)
}
