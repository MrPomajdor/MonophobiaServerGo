package game

import (
	"MonophobiaServer/internal/data"
	"MonophobiaServer/internal/entity"
)

type Player struct {
	ID               int32
	Transforms       data.Transforms
	FutureTransforms data.Transforms
	Inputs           data.Inputs

	Lobby *Lobby

	UserData *entity.UserData
}

func (p *Player) SendPacket(over entity.Network, packet *entity.Packet) {
	packet.Network = over
	respPackage := &MessagePackage{packet, p}
	*(p.Lobby.Manager.ResponsePacketChannel) <- respPackage
}
