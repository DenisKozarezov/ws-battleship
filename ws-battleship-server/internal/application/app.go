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
	cfg        *config.AppConfig
	httpServer *http.Server
	wsListener *handlers.WebsocketListener
	logger     logger.Logger

	mu     sync.RWMutex
	joinCh chan *domain.Player
	rooms  map[string]*domain.Room
}

func NewApp(cfg *config.AppConfig, logger logger.Logger) *App {
	joinCh := make(chan *domain.Player, cfg.ClientsConnectionsMax)

	return &App{
		cfg:        cfg,
		logger:     logger,
		wsListener: handlers.NewWebsocketListener(cfg, logger, joinCh),
		joinCh:     joinCh,
		rooms:      make(map[string]*domain.Room, cfg.ClientsConnectionsMax),
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

	a.logger.Infof("starting a server :%s", a.cfg.Port)
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
	a.logger.Infof("server :%s is gracefully shutdown", a.cfg.Port)
}

func (a *App) Shutdown() error {
	a.logger.Info("shutting the server down...")

	a.wsListener.Close()
	for _, room := range a.rooms {
		if err := room.Close(); err != nil {
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
					r.logger.Errorf("failed to connect a player to free room: %s", err)
				}
			}
		}
	}
}

func (r *App) connectPlayerToFreeRoom(ctx context.Context, newPlayer *domain.Player) error {
	room := r.findFreeRoom()
	if room == nil {
		room = r.createNewRoom(ctx)
	}

	if err := room.JoinNewPlayer(newPlayer); err != nil {
		return err
	}

	return nil
}

func (r *App) findFreeRoom() *domain.Room {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.rooms) == 0 {
		return nil
	}

	for _, room := range r.rooms {
		if !room.IsFull() {
			return room
		}
	}
	return nil
}

func (r *App) createNewRoom(ctx context.Context) *domain.Room {
	room := domain.NewRoom(ctx, r.cfg, r.logger)

	r.mu.Lock()
	r.rooms[room.ID()] = room
	r.mu.Unlock()

	r.logger.Infof("new room with id=%s was created [rooms: %d]", room.ID(), len(r.rooms))
	return room
}
