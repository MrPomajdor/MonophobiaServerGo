package state

import "MonophobiaServer/internal/entity"

type GameState struct {
	Players      []*entity.LogicPlayer
	GameEntities []*entity.GameEntity
}
