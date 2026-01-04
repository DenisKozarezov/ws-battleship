package domain

import (
	"net/http"

	"github.com/google/uuid"
)

type PlayerModel struct {
	Board    Board
	ID       string
	Nickname string
}

func (m *PlayerModel) Equal(rhs *PlayerModel) bool {
	if rhs == nil {
		return false
	}
	return m.ID == rhs.ID
}

func NewPlayerModel(board Board, metadata ClientMetadata) *PlayerModel {
	return &PlayerModel{
		Board:    board,
		ID:       metadata.ClientID,
		Nickname: metadata.Nickname,
	}
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
