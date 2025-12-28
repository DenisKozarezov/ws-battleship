package models

import (
	"math/rand"
	"ws-battleship-shared/domain"
)

type Player struct {
	Board    domain.Board
	Nickname string
}

func NewPlayer(nickname string) *Player {
	return &Player{
		Board:    shuffleBoard(),
		Nickname: nickname,
	}
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
