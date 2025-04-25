package data

type Vector3 struct {
	X float32
	Y float32
	Z float32
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

type ItemData struct {
	ItemID     int32
	Transforms Transforms
}
