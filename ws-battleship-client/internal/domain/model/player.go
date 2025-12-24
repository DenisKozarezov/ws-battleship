package model

type Player struct {
	Board    Board
	Nickname string
}

func NewPlayer(nickname string) *Player {
	var board = Board{
		{Miss, Miss, Miss, 0, Alive, Alive, Dead, 0, 0, 0},
		{Miss, Miss, 0, 0, 0, Miss, 0, 0, Miss, 0},
		{Miss, Alive, 0, 0, 0, 0, 0, 0, Miss, 0},
		{0, Alive, 0, 0, Dead, 0, 0, Miss, 0, 0},
		{0, 0, 0, 0, Dead, 0, 0, 0, 0, 0},
		{0, Miss, Miss, 0, Alive, 0, 0, 0, Miss, 0},
		{0, Miss, 0, 0, Alive, 0, 0, Miss, 0, 0},
		{0, 0, 0, 0, 0, 0, 0, Alive, 0, 0},
		{0, Miss, Miss, Miss, 0, 0, 0, 0, 0, Miss},
		{Miss, Miss, Dead, 0, 0, Dead, 0, Dead, Miss, Miss},
	}

	return &Player{
		Board:    board,
		Nickname: nickname,
	}
}
