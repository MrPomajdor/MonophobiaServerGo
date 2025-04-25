package entity

type UserData struct {
	Name      string
	Cosmetics []string
	Skin      string
	SteamID   string
}

type LogicPlayer struct {
	ID            int32
	IP            string
	UserData      *UserData
	NetworkClient *Client
}

type NetworkPlayerInfo struct {
	ID        int32
	Name      string
	Cosmetics []string
	Skin      string
	IsMonster bool
	IsHost    bool
}

func NewPlayer() *LogicPlayer {
	pl := &LogicPlayer{}
	pl.UserData = &UserData{}
	pl.ID = -1
	return pl
}
