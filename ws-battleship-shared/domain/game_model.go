package domain

type GameModel struct {
	LeftPlayer  *PlayerModel
	RightPlayer *PlayerModel
}

func (g GameModel) Copy() GameModel {
	var leftPlayer *PlayerModel
	if g.LeftPlayer != nil {
		leftPlayer = NewPlayerModel(g.LeftPlayer.Board, ClientMetadata{
			ClientID: g.LeftPlayer.ID,
			Nickname: g.LeftPlayer.Nickname,
		})
	}

	var rightPlayer *PlayerModel
	if g.RightPlayer != nil {
		rightPlayer = NewPlayerModel(g.RightPlayer.Board, ClientMetadata{
			ClientID: g.RightPlayer.ID,
			Nickname: g.RightPlayer.Nickname,
		})
	}

	return GameModel{
		LeftPlayer:  leftPlayer,
		RightPlayer: rightPlayer,
	}
}
