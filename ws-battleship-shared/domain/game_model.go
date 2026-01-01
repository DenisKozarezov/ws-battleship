package domain

type GameModel struct {
	LeftPlayer  *PlayerModel
	RightPlayer *PlayerModel
	Messages    []string
}
