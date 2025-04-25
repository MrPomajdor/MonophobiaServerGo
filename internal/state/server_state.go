package state

import (
	"MonophobiaServer/internal/entity"
	"math/rand/v2"
	"slices"
)

type ServerState struct {
	Clients     []*entity.Client
	ClientIDMap map[int32]*entity.Client
}

func (s *ServerState) InitializePlayer(name string, client *entity.Client) *entity.LogicPlayer {
	NewPlayer := entity.NewPlayer()
	NewPlayer.Name = name
	NewPlayer.IP = client.IP
	NewPlayer.NetworkClient = client
	client.LogicPlayer = NewPlayer
	idValid := true
	for {

		NewPlayer.ID = rand.Int32()
		for _, v := range s.Clients {
			if (v.LogicPlayer != nil) && (v.LogicPlayer.ID == NewPlayer.ID) {
				idValid = false
				break
			}
			idValid = true
		}
		if idValid {
			break
		}
	}
	s.Clients = append(s.Clients, client)
	s.ClientIDMap[NewPlayer.ID] = client
	return NewPlayer
}

func (s *ServerState) RemoveClient(cl *entity.Client) {
	delete(s.ClientIDMap, cl.LogicPlayer.ID)

	s.Clients = slices.DeleteFunc(s.Clients, func(n *entity.Client) bool {
		return n == cl
	})
}
