package GameServer

import (
	"time"

	"MonophobiaServer/messages"
	// initrd.img"github.com/deeean/go-vector/vector3"
)

type MapsStruct struct {
	Lobby string
	Grid  string
}

var Maps MapsStruct = MapsStruct{Lobby: "lobby0", Grid: "grid0"}

type LobbyMessage int

const (
	LobbyShutdown LobbyMessage = iota
	LobbyRefresh
)

type WorldState struct {
	Items []Item
}

type Vector3 struct {
	X float32
	Y float32
	Z float32
}

/*
func (v *Vector3) Distance(v2 *Vector3) float64 {
	return math.Sqrt((math.Pow(v2.X-v.X, 2) + math.Pow(v2.Y-v.Y, 2) + math.Pow(v2.Z-v.Z, 2)))
}*/

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
	MoveDirection Vector3
}
type Transforms struct {
	Position            Vector3
	Rotation            Vector3
	RealVelocity        Vector3
	RealAngularVelocity Vector3
}

type PlayerData struct {
	PlayerID   int32
	Transforms Transforms
	Inputs     Inputs
}

type Lobby struct {
	Owner             *Player
	Name              string
	Players           []*Player
	Map               string
	MapSeed           int32
	MaxPlayers        int32
	PasswordProtected bool
	Password          string
	ID                int32
	Started           bool
	WorldState        WorldState
	LogicChannel      chan Packet
	MessageChannel    chan LobbyMessage
	TickRate          time.Duration
}

func (l *Lobby) ToNetwork() *NetworkLobbyInfo {
	inf := &NetworkLobbyInfo{}
	inf.LobbyName = l.Name
	inf.MapName = l.Map
	inf.Time = 0
	inf.Players = make([]NetworkPlayerInfo, len(l.Players))
	for i, pl := range l.Players {
		inf.Players[i] = *pl.ToNetwork()
	}
	return inf
}

type NetworkLobbyInfo struct {
	LobbyName string
	MapName   string
	Time      int32
	Players   []NetworkPlayerInfo
}

type Player struct {
	ID               int32
	Name             string
	IP               string
	PlayerData       PlayerData
	Cosmetics        []string
	Skin             string
	IsMonster        bool
	IsHost           bool
	SteamID          string
	Lobby            *Lobby
	NetworkClient    *Client
	FutureTransforms Transforms
	Transforms       Transforms
}

func (pl *Player) ToNetwork() *NetworkPlayerInfo {
	newData := &NetworkPlayerInfo{}
	newData.Name = pl.Name
	newData.ID = pl.ID
	newData.Cosmetics = pl.Cosmetics
	newData.IsHost = pl.IsHost
	newData.IsMonster = pl.IsMonster
	return newData
}

type NetworkPlayerInfo struct {
	ID        int32
	Name      string
	Cosmetics []string
	Skin      string
	IsMonster bool
	IsHost    bool
}

type Packet struct {
	Client         *Client
	Header         messages.Header
	Flag           messages.Flag
	FullMsgLen     int32
	Payload        []byte
	payloadPointer int32
}
