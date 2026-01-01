package domain

import (
	"math/rand"
	"net/http"

	"github.com/google/uuid"
)

type PlayerModel struct {
	Board    Board
	ID       string
	Nickname string
}

func NewPlayerModel(metadata ClientMetadata) *PlayerModel {
	return &PlayerModel{
		Board:    shuffleBoard(),
		ID:       metadata.ClientID,
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

type ClientID = string

type ClientMetadata struct {
	ClientID ClientID
	Nickname string
}

func NewClientMetadata(nickname string) ClientMetadata {
	return ClientMetadata{
		ClientID: uuid.New().String(),
		Nickname: nickname,
	}
}

func ParseClientMetadataToHeaders(metadata ClientMetadata) http.Header {
	headers := make(http.Header)
	headers.Set("X-Client-ID", metadata.ClientID)
	headers.Set("X-Nickname", metadata.Nickname)
	return headers
}

func ParseClientMetadataFromHeaders(r *http.Request) ClientMetadata {
	return ClientMetadata{
		ClientID: r.Header.Get("X-Client-ID"),
		Nickname: r.Header.Get("X-Nickname"),
	}
}
