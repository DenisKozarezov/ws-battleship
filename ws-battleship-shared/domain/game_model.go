package domain

type GameModel struct {
	TurnCount int
	Players   map[string]*PlayerModel
}
