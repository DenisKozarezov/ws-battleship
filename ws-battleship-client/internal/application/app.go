package application

import (
	"context"
	"errors"
	"sync"
	"time"
	"ws-battleship-client/internal/application/states"
	"ws-battleship-client/internal/config"
	"ws-battleship-client/internal/domain/views"
	"ws-battleship-shared/domain"
	"ws-battleship-shared/events"
	"ws-battleship-shared/pkg/logger"

	tea "github.com/charmbracelet/bubbletea"
)

type Client interface {
	Messages() <-chan events.Event
	Connect(ctx context.Context, metadata domain.ClientMetadata) error
	SendMessage(e events.Event) error
	Shutdown() error
}

type App struct {
	ctx context.Context

	renderCh chan tea.Model

	cfg          *config.Config
	logger       logger.Logger
	stateMachine states.StateMachine
	mainMenu     *views.MainMenuView
	metadata     domain.ClientMetadata
}

func NewApp(ctx context.Context, cfg *config.Config, logger logger.Logger) *App {
	stateMachine := states.NewStateMachine()

	app := &App{
		ctx:          ctx,
		renderCh:     make(chan tea.Model, 1),
		cfg:          cfg,
		logger:       logger,
		stateMachine: stateMachine,
		mainMenu:     views.NewMainMenuView(),
	}

	stateMachine.SetStateSwitchedHandler(app.onApplicationStateSwitched)

	return app
}

func (a *App) Run(ctx context.Context) {
	var wg sync.WaitGroup
	a.runGameLoop(ctx, &wg)
	a.runRenderLoop(ctx, &wg)

	a.stateMachine.SwitchState(states.NewMainMenuState(a.stateMachine, a.logger))

	<-ctx.Done()
	a.logger.Info("received a signal to shutdown the client")
	wg.Wait()

	if err := a.Shutdown(); err != nil {
		a.logger.Fatalf("failed to shutdown a client: %s", err)
	}
	a.logger.Info("client is gracefully shutdown")
}

func (a *App) Shutdown() error {
	a.logger.Info("shutting the client down...")
	return nil
}

func (a *App) runGameLoop(ctx context.Context, wg *sync.WaitGroup) {
	const fps = 60
	const fixedTime = time.Second / fps

	wg.Add(1)
	go func() {
		ticker := time.NewTicker(fixedTime)
		defer func() {
			wg.Done()
			ticker.Stop()
		}()

		for {
			if err := ctx.Err(); err != nil {
				return
			}

			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.stateMachine.FixedUpdate()
			}
		}
	}()
}

func (a *App) runRenderLoop(ctx context.Context, wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
			close(a.renderCh)
		}()

		var currentProgram *tea.Program

		for {
			select {
			case <-ctx.Done():
				if currentProgram != nil {
					currentProgram.Quit()
				}
				return

			case currentView := <-a.renderCh:
				if currentView == nil {
					return
				}

				if currentProgram != nil {
					currentProgram.Quit()
					time.Sleep(50 * time.Millisecond)
				}

				clearTerminal()

				currentProgram = tea.NewProgram(currentView, tea.WithContext(ctx), tea.WithAltScreen())
				go func(p *tea.Program) {
					if _, err := p.Run(); err != nil {
						if !errors.Is(err, tea.ErrProgramKilled) {
							a.logger.Errorf("failed to render view: %s", err)
						}
					}
				}(currentProgram)
			}
		}
	}()
}

func (a *App) onApplicationStateSwitched(currentView tea.Model) {
	a.renderCh <- currentView
}
