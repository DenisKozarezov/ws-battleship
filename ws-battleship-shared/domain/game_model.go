package domain

type GameModel struct {
	Players  map[string]*PlayerModel
	Messages []string
}

func NewGameModel(players map[string]*PlayerModel) GameModel {
	return GameModel{
		Players: players,
	}
}
