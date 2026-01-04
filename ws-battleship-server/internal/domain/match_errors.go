package domain

import "errors"

var (
	ErrInvalidTarget = errors.New("invalid target")
	ErrNotYourTurn   = errors.New("this player doesn't have permission to fire")
)
