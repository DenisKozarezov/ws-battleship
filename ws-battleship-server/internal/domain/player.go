package domain

import (
	"fmt"
	"ws-battleship-shared/domain"
)

type Player struct {
	Client
	Model *domain.PlayerModel
}

func NewPlayer(client Client, metadata domain.ClientMetadata) *Player {
	return &Player{
		Model:  domain.NewPlayerModel(metadata),
		Client: client,
	}
}

func (p *Player) String() string {
	return fmt.Sprintf(`'%s' [%s]`, p.Nickname(), p.ID())
}

func (p *Player) Nickname() string {
	return p.Model.Nickname
}
