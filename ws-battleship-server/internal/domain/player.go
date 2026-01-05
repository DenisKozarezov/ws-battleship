package domain

import (
	"fmt"
	"math/rand"
	"ws-battleship-shared/domain"
)

type Player struct {
	Client
	Model *domain.PlayerModel
}

func NewPlayer(client Client, metadata domain.ClientMetadata) *Player {
	return &Player{
		Model:  domain.NewPlayerModel(randomizeBoard(), metadata),
		Client: client,
	}
}

func (p *Player) String() string {
	return fmt.Sprintf(`'%s' [%s]`, p.Nickname(), p.ID())
}

func (p *Player) Nickname() string {
	return p.Model.Nickname
}

func randomizeBoard() domain.Board {
	ships := [...]int{4, 3, 3, 2, 2, 2, 1, 1, 1, 1}

	for {
		var board domain.Board
		ok := true

		for _, size := range ships {
			placed := false

			for attempts := 0; attempts < 1000; attempts++ {
				x := rand.Intn(board.Size())
				y := rand.Intn(board.Size())
				horizontal := rand.Intn(2) == 0

				if canPlace(&board, x, y, size, horizontal) {
					placeShip(&board, x, y, size, horizontal)
					placed = true
					break
				}
			}

			if !placed {
				ok = false
				break
			}
		}

		if ok {
			return board
		}
	}
}

func canPlace(
	board *domain.Board,
	x, y int,
	length int,
	horizontal bool,
) bool {
	for i := 0; i < length; i++ {
		nx, ny := x, y
		if horizontal {
			ny += i
		} else {
			nx += i
		}

		if nx < 0 || nx >= board.Size() || ny < 0 || ny >= board.Size() {
			return false
		}

		for dx := -1; dx <= 1; dx++ {
			for dy := -1; dy <= 1; dy++ {
				cx := nx + dx
				cy := ny + dy

				if cx >= 0 && cx < board.Size() && cy >= 0 && cy < board.Size() {
					if board[cx][cy] == domain.Ship {
						return false
					}
				}
			}
		}
	}
	return true
}

func placeShip(
	board *domain.Board,
	x, y int,
	length int,
	horizontal bool,
) {
	for i := 0; i < length; i++ {
		if horizontal {
			board[x][y+i] = domain.Ship
		} else {
			board[x+i][y] = domain.Ship
		}
	}
}
