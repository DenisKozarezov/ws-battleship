package domain

type GameModel struct {
	Player1  *Player
	Player2  *Player
	Messages []string
}

func NewGameModel() *GameModel {
	return &GameModel{
		Player1: NewPlayer("Player 1"),
		Player2: NewPlayer("Player 2"),
	}
}
