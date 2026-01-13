package domain

import (
	"fmt"
	"ws-battleship-server/internal/delivery/websocket"
	"ws-battleship-shared/domain"
)

type VisibleCell struct {
	X, Y byte
}

type Player struct {
	websocket.Client
	Model      *domain.PlayerModel
	visibility []VisibleCell
}

func NewPlayer(client websocket.Client, metadata domain.ClientMetadata) *Player {
	model := domain.NewPlayerModel(domain.RandomizeBoard(), metadata)
	return &Player{
		Model:  model,
		Client: client,
	}
}

func (p *Player) Equal(rhs *Player) bool {
	if rhs == nil {
		return false
	}
	return p.Model.Equal(rhs.Model)
}

func (p *Player) Compare(rhs *Player) int {
	if rhs == nil {
		return -1
	}
	return p.Model.Compare(rhs.Model)
}

func (p *Player) String() string {
	return fmt.Sprintf(`'%s' [%s]`, p.Nickname(), p.ID())
}

func (p *Player) Nickname() string {
	return p.Model.Nickname
}

func (p *Player) RevealCell(cellX, cellY byte) {
	p.visibility = append(p.visibility, VisibleCell{X: cellX, Y: cellY})
}

func (p *Player) maskBoardForPlayer(targetPlayer *Player) domain.Board {
	if targetPlayer == nil {
		return p.Model.Board
	}

	var copiedBoard domain.Board
	for i := 0; i < len(targetPlayer.visibility); i++ {
		visibleX := targetPlayer.visibility[i].X
		visibleY := targetPlayer.visibility[i].Y
		copiedBoard.SetCell(visibleX, visibleY, p.Model.Board.GetCellType(visibleX, visibleY))
	}

	return copiedBoard
}
