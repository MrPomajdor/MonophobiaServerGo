package game

import (
	"MonophobiaServer/internal/data"
	"MonophobiaServer/internal/entity"
	"MonophobiaServer/internal/messages"
)

func (g *Game) checkPlayerMovement() {
	var updatedPlayersPos []data.PlayerData

	for _, pl := range g.Players {
		if pl.Transforms != pl.FutureTransforms {
			// TODO : some calculations
			pl.Transforms = pl.FutureTransforms
			elem := data.PlayerData{PlayerID: pl.ID, Transforms: pl.Transforms, Inputs: pl.Inputs}
			updatedPlayersPos = append(updatedPlayersPos, elem)
		}
	}

	if len(updatedPlayersPos) != 0 && len(g.Players) > 1 {
		var playersPosUpdatePacket struct {
			Players []data.PlayerData
		}
		playersPosUpdatePacket.Players = updatedPlayersPos

		pac := entity.Packet{}
		pac.Header = messages.Data
		pac.Flag = messages.Response.PlayerTransforms
		pac.AddToPayload(&playersPosUpdatePacket)
		// log.Debug(err)
		for _, pl := range g.Players {
			pl.SendPacket(entity.NET_TCP, &pac)
		}
	}

}
