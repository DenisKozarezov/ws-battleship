package application

import (
	"context"
	"net"
	"net/http"
	"sync"
	"ws-battleship-server/internal/config"
	"ws-battleship-server/internal/delivery/http/routers"
	"ws-battleship-server/internal/delivery/websocket/handlers"
	"ws-battleship-server/internal/domain"
	"ws-battleship-shared/pkg/logger"
)

type App struct {
	cfg        *config.Config
	httpServer *http.Server
	wsListener *handlers.WebsocketListener
	logger     logger.Logger

	mu      sync.RWMutex
	joinCh  chan *domain.Player
	matches map[string]*domain.Match
}

func NewApp(cfg *config.Config, logger logger.Logger) *App {
	joinCh := make(chan *domain.Player, cfg.App.ClientsConnectionsMax)
	return &App{
		cfg:        cfg,
		logger:     logger,
		wsListener: handlers.NewWebsocketListener(&cfg.App, logger, joinCh),
		joinCh:     joinCh,
		matches:    make(map[string]*domain.Match, cfg.App.ClientsConnectionsMax),
	}
}

func (a *App) Run(ctx context.Context, router routers.Router) {
	a.SetupRoutes(router)

	a.httpServer = &http.Server{
		Addr:           ":" + a.cfg.App.Port,
		Handler:        router,
		MaxHeaderBytes: 1 << 10,
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}

	a.logger.Infof("starting a server :%s", a.cfg.App.Port)
	go func() {
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.logger.Fatalf("failed to run a server: %s", err)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		a.handleConnections(ctx)
	}()

	<-ctx.Done()
	a.logger.Info("received a signal to shutdown the server")
	wg.Wait()

	if err := a.Shutdown(); err != nil {
		a.logger.Fatalf("failed to shutdown a server: %s", err)
	}
	a.logger.Infof("server :%s is gracefully shutdown", a.cfg.App.Port)
}

func (a *App) Shutdown() error {
	a.logger.Info("shutting the server down...")

	a.wsListener.Close()
	for _, match := range a.matches {
		if err := match.Close(); err != nil {
			a.logger.Errorf("failed to close a room: %s", err)
		}
	}

	return a.httpServer.Close()
}

func (a *App) SetupRoutes(router routers.Router) {
	router.GET("/ws", a.wsListener.HandleWebsocketConnection)
}

func (r *App) handleConnections(ctx context.Context) {
	defer close(r.joinCh)

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		// Register incoming clients, when they establish a connection.
		case newPlayer, opened := <-r.joinCh:
			if opened {
				if err := r.connectPlayerToFreeRoom(ctx, newPlayer); err != nil {
					r.logger.Errorf("failed to connect a player to free match: %s", err)
				}
			}
		}
	}
}

func (r *App) connectPlayerToFreeRoom(ctx context.Context, newPlayer *domain.Player) error {
	match := r.findFreeMatch()
	if match == nil {
		match = r.createNewMatch(ctx)
	}

	return match.JoinNewPlayer(newPlayer)
}

func (r *App) findFreeMatch() *domain.Match {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.matches) == 0 {
		return nil
	}

	for _, match := range r.matches {
		if match.CheckIsAvailableForJoin() == nil {
			return match
		}
	}
	return nil
}

func (r *App) createNewMatch(ctx context.Context) *domain.Match {
	match := domain.NewMatch(ctx, r.cfg, r.logger)

	r.mu.Lock()
	r.matches[match.ID()] = match
	r.mu.Unlock()

	r.logger.Infof("new match with id=%s was created [rooms: %d]", match.ID(), len(r.matches))
	return match
}
