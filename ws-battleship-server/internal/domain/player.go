package domain

import (
	"fmt"
	"math/rand"
	"ws-battleship-shared/domain"
)

type Player struct {
	Client
	Model *domain.PlayerModel
}

func NewPlayer(client Client, metadata domain.ClientMetadata) *Player {
	return &Player{
		Model:  domain.NewPlayerModel(shuffleBoard(), metadata),
		Client: client,
	}
}

func (p *Player) String() string {
	return fmt.Sprintf(`'%s' [%s]`, p.Nickname(), p.ID())
}

func (p *Player) Nickname() string {
	return p.Model.Nickname
}

func shuffleBoard() domain.Board {
	var b domain.Board

	cells := []domain.CellType{domain.Empty, domain.Dead, domain.Alive, domain.Miss}

	for i := 0; i < b.Size(); i++ {
		for j := 0; j < b.Size(); j++ {
			r := rand.Intn(len(cells))
			b[i][j] = cells[r]
		}
	}

	return b
}
