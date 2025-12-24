package application

import (
	"fmt"
)

type Game struct {
	Player1 *Player
	Player2 *Player
}

func NewGame() *Game {
	return &Game{
		Player1: NewPlayer(),
		Player2: NewPlayer(),
	}
}

func (g *Game) RenderScreen() {
	clearTerminal()

	b1Lines := g.Player1.Board.Lines()
	b2Lines := g.Player2.Board.Lines()
	for i := range b1Lines {
		fmt.Println(b1Lines[i] + "        " + b2Lines[i])
	}
}
