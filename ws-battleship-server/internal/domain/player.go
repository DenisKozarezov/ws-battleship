package domain

import (
	"fmt"
	"ws-battleship-shared/domain"
)

type Player struct {
	*domain.PlayerModel
	*Client
}

func NewPlayer(client *Client) *Player {
	return &Player{
		PlayerModel: domain.NewPlayerModel(client.metadata),
		Client:      client,
	}
}

func (p *Player) String() string {
	return fmt.Sprintf(`'%s' [%s]`, p.Nickname(), p.ID())
}

func (p *Player) Nickname() string {
	return p.PlayerModel.Nickname
}
