package GameServer

import (
	"MonophobiaServer/messages"

	"github.com/deeean/go-vector/vector3"
)

type WorldState struct {
	Items []Item
}
type Item struct {
	ID         int32
	Name       string
	Activated  bool
	Transforms Transforms
}
type Inputs struct {
	IsSprinting   bool
	IsMoving      bool
	IsCrouching   bool
	MoveDirection vector3.Vector3
}
type Transforms struct {
	Position            vector3.Vector3
	Rotation            vector3.Vector3
	RealVelocity        vector3.Vector3
	RealAngularVelocity vector3.Vector3
}
type PlayerData struct {
	Transforms Transforms
	Inputs     Inputs
}

type Lobby struct {
	Owner             *Player
	Name              string
	Players           []*Player
	Map               string
	MapSeed           int32
	MaxPlayers        uint32
	PasswordProtected bool
	Password          string
	ID                int32
	Started           bool
	WorldState        WorldState
}
type Player struct {
	ID         int32
	Name       string
	IP         string
	PlayerData PlayerData
	Cosmetics  []string
	Skin       string
	IsMonster  bool
	IsHost     bool
	Lobby      *Lobby
}

type Packet struct {
	Header         messages.Header
	Flag           messages.Flag
	FullMsgLen     uint32
	Payload        []byte
	payloadPointer uint32
}
