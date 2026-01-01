package domain

import (
	"math/rand"
	"net/http"
)

type PlayerModel struct {
	Board    Board
	ID       string
	Nickname string
}

func NewPlayerModel(playerID string, metadata ClientMetadata) *PlayerModel {
	return &PlayerModel{
		Board:    shuffleBoard(),
		ID:       playerID,
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

type ClientMetadata struct {
	Nickname string
}

func ParseClientMetadataToHeaders(metadata ClientMetadata) http.Header {
	headers := make(http.Header)
	headers.Set("X-Nickname", metadata.Nickname)
	return headers
}

func ParseClientMetadataFromHeaders(r *http.Request) ClientMetadata {
	return ClientMetadata{
		Nickname: r.Header.Get("X-Nickname"),
	}
}
