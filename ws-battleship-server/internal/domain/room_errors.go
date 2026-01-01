package domain

import "errors"

var (
	ErrRoomIsClosed        = errors.New("room is closed")
	ErrRoomIsFull          = errors.New("room's capacity is exceeded")
	ErrPlayerAlreadyInRoom = errors.New("player is already in room")
	ErrPlayerNotExist      = errors.New("player doesn't exist")
	ErrAlreadyStarted      = errors.New("already started")
)
