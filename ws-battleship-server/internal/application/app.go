package application

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"
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

	mu      sync.RWMutex
	clients map[string]*domain.Client
}

func NewApp(ctx context.Context, cfg *config.AppConfig, logger logger.Logger) *App {
	return &App{
		cfg:        cfg,
		logger:     logger,
		wsListener: handlers.NewWebsocketListener(ctx, cfg, logger),
		clients:    make(map[string]*domain.Client, cfg.ClientsConnectionsMax),
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
	wg.Add(2)
	go func() {
		defer wg.Done()
		a.handleConnections(ctx)
	}()
	go func() {
		defer wg.Done()
		a.pingClients(ctx)
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
	for _, client := range a.clients {
		if err := a.UnregisterClient(client); err != nil {
			a.logger.Errorf("failed to unregister a client: %s", err)
		}
	}
	a.wsListener.WaitForAllConnections()

	return a.httpServer.Close()
}

func (a *App) SetupRoutes(router routers.Router) {
	router.GET("/ws", a.wsListener.HandleWebsocketConnection)
}

func (a *App) handleConnections(ctx context.Context) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		// Register incoming clients, when they establish a connection.
		case newClient, opened := <-a.wsListener.JoinChan():
			if opened {
				a.RegisterNewClient(newClient)
			}

		case msg, opened := <-a.wsListener.Messages():
			if opened {
				a.handleMessage(msg)
			}
		}
	}
}

func (a *App) handleMessage(event domain.Event) {
	a.logger.Debug("Event Type: %d; Timestamp: %s; Payload: %s", event.Type, event.Timestamp, string(event.Data))
}

func (a *App) pingClients(ctx context.Context) {
	pingTicker := time.NewTicker(a.cfg.KeepAlivePeriod)
	defer pingTicker.Stop()

	deadClients := make(chan *domain.Client, a.cfg.ClientsConnectionsMax)
	defer close(deadClients)

	for {
		if err := ctx.Err(); err != nil {
			return
		}

		select {
		case <-ctx.Done():
			return

		// We should periodically send a ping-message to all clients just to be ensured, that the clients
		// are still alive. If no, then the server collects dead clients to a special queue for further
		// unregistering.
		case <-pingTicker.C:
			a.mu.RLock()
			for _, client := range a.clients {
				go a.pingClient(client, deadClients)
			}
			a.mu.RUnlock()

		// We must kick potentially dead clients who didn't response to our ping-message. There are literally zero
		// reasons to keep stalled connections alive, so the server deallocates them for other needs.
		case deadClient := <-deadClients:
			a.logger.Infof("client %s didn't response to ping and was declared as potentially dead by the server, unregistering it...", deadClient.String())
			if err := a.UnregisterClient(deadClient); err != nil {
				a.logger.Errorf("failed to disconnect a dead client: %s", err)
			}
		}
	}
}

func (a *App) pingClient(client *domain.Client, deadClients chan<- *domain.Client) {
	if err := client.Ping(); err != nil {
		a.logger.Errorf("failed to ping a client id=%s: %s", client.ID(), err)
		deadClients <- client
	}
}

func (a *App) RegisterNewClient(newClient *domain.Client) {
	a.mu.Lock()
	a.clients[newClient.ID()] = newClient
	a.mu.Unlock()

	a.logger.Infof("client %s is now connected", newClient.String())
}

func (a *App) UnregisterClient(client *domain.Client) error {
	client.Close()

	a.mu.Lock()
	delete(a.clients, client.ID())
	a.mu.Unlock()

	a.logger.Infof("client id=%s was unregistered", client.ID())

	return nil
}

func (a *App) Broadcast(eventType domain.EventType, obj any) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	for _, client := range a.clients {
		if err := client.SendMessage(eventType, obj); err != nil {
			a.logger.Errorf("failed to send a broadcast message to client id=%s", client.ID())
		}
	}
}
