package domain

import "errors"

var (
	ErrRoomIsFull          = errors.New("room's capacity is exceeded")
	ErrPlayerAlreadyInRoom = errors.New("player is already in room")
	ErrPlayerNotExist      = errors.New("player doesn't exist")
)
