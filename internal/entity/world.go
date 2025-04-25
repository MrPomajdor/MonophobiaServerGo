package entity

type NetworkWorldState struct {
	Items []GameEntity
}
type WorldState struct {
	Items map[int32]*GameEntity
}

func (st *WorldState) ToNetwork() *NetworkWorldState {
	res := NetworkWorldState{}
	res.Items = make([]GameEntity, len(st.Items))
	for i, it := range st.Items {
		res.Items[i] = *it
	}
	return &res
}

func (st *WorldState) InputFromNetwork(ns *NetworkWorldState) {
	st.Items = make(map[int32]*GameEntity)
	for _, itm := range ns.Items {
		newItmPointer := GameEntity(itm)
		st.Items[itm.ID] = &newItmPointer
	}
}
