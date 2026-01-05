package domain

import (
	"net/http"
	"ws-battleship-shared/pkg/math"

	"github.com/google/uuid"
)

type PlayerModel struct {
	Board     Board
	ID        string
	Nickname  string
	ShipCells byte
}

func NewPlayerModel(board Board, metadata ClientMetadata) *PlayerModel {
	var shipCells byte
	for i := 0; i < board.Size(); i++ {
		for j := 0; j < board.Size(); j++ {
			if board.GetCellType(byte(i), byte(j)) == Ship {
				shipCells++
			}
		}
	}

	return &PlayerModel{
		Board:     board,
		ID:        metadata.ClientID,
		Nickname:  metadata.Nickname,
		ShipCells: shipCells,
	}
}

func (m *PlayerModel) Equal(rhs *PlayerModel) bool {
	if rhs == nil {
		return false
	}
	return m.ID == rhs.ID
}

func (m *PlayerModel) IsDead() bool {
	return m.ShipCells == 0
}

func (m *PlayerModel) DecrementCell() {
	m.ShipCells = math.Clamp(m.ShipCells-1, 0, byte(m.Board.Size()))
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
