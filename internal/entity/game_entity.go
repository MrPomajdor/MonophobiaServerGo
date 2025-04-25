package entity

import "MonophobiaServer/internal/data"

type GameEntity struct {
	ID               int32
	Name             string
	Activated        bool
	Transforms       data.Transforms
	FutureTransforms data.Transforms
}
