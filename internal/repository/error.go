package repository

import "errors"

var ErrDuplicateEmail = errors.New("email already exists")

var ErrDuplicateGameCode = errors.New("gameCode already exists")

var ErrGameNotFound = errors.New("game ID not exists")

var ErrStageNotFound = errors.New("Stage ID not exists")

var ErrLocationNotFound = errors.New("Location ID not exists")

var (
	ErrNextStageNotFound      = errors.New("next stage not found")
	ErrDuplicateNextStageSort = errors.New(
		"sort_id already exists for this game and stage",
	)
)
