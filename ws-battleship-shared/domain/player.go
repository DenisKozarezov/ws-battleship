package domain

import (
	"math/rand"
	"ws-battleship-shared/events"
)

type PlayerModel struct {
	Board    Board
	Nickname string
}

func NewPlayerModel(metadata events.ClientMetadata) *PlayerModel {
	return &PlayerModel{
		Board:    shuffleBoard(),
		Nickname: metadata.Nickname,
	}
}

func shuffleBoard() Board {
	var b Board

	cells := []CellType{Empty, Dead, Alive, Miss}

	for i := 0; i < b.Size(); i++ {
		for j := 0; j < b.Size(); j++ {
			r := rand.Intn(len(cells))
			b[i][j] = cells[r]
		}
	}

	return b
}
