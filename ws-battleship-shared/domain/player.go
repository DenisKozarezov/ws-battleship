package domain

import "math/rand"

type Player struct {
	Board    Board
	Nickname string
}

func NewPlayer(nickname string) *Player {
	return &Player{
		Board:    shuffleBoard(),
		Nickname: nickname,
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
