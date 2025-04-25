package messages

type Header uint16

const (
	Ack           Header = 0x0000
	Echo          Header = 0x0100
	Hello         Header = 0x0200
	Data          Header = 0x0300
	Disconnecting Header = 0x0400
	Rejected      Header = 0xFFFF
	ImHere        Header = 0xAAAA // imagine the client screaming over the twisted pair
)

type Flag uint8

const (
	None Flag = 0x00
)

type RequestStruct struct {
	PlayerList       Flag
	LobbyList        Flag
	WorldState       Flag
	ItemList         Flag
	NetworkVariables Flag
}

var Request = RequestStruct{
	PlayerList:       0x04,
	LobbyList:        0x07,
	WorldState:       0xD0,
	ItemList:         0x1A,
	NetworkVariables: 0xEB,
}

type PostStruct struct {
	JoinLobby              Flag
	CreateLobby            Flag
	UpdateLobbyInfo        Flag
	PlayerTransformData    Flag
	LobbyInfo              Flag
	WorldState             Flag
	ItemPickup             Flag
	ItemDrop               Flag
	InventorySwitch        Flag
	StartMap               Flag
	ItemIntInf             Flag
	Voice                  Flag
	InteractableMessage    Flag
	CodeInteractionMessage Flag
	Transform              Flag
	NetworkVarSync         Flag
	ChatMessage            Flag
}

var Post = PostStruct{
	JoinLobby:              0x08,
	CreateLobby:            0x11,
	UpdateLobbyInfo:        0x10,
	PlayerTransformData:    0xA0,
	LobbyInfo:              0xA1,
	WorldState:             0xA2,
	ItemPickup:             0xA5,
	ItemDrop:               0xA6,
	InventorySwitch:        0xA7,
	StartMap:               0xA8,
	ItemIntInf:             0xA4,
	Voice:                  0xAC,
	InteractableMessage:    0xAD,
	CodeInteractionMessage: 0xAE,
	Transform:              0xAF,
	NetworkVarSync:         0xBE,
	ChatMessage:            0xE1,
}

type ResponseStruct struct {
	IDAssign               Flag
	PlayerList             Flag
	LobbyList              Flag
	Error                  Flag
	ClosingCon             Flag
	LobbyListChanged       Flag
	LobbyClosing           Flag
	PlayerTransforms       Flag
	LobbyInfo              Flag
	WorldState             Flag
	ItemIntInf             Flag
	ItemList               Flag
	ItemPickup             Flag
	ItemDrop               Flag
	InventorySwitch        Flag
	StartMap               Flag
	InteractableMessage    Flag
	CodeInteractionMessage Flag
	PlayerData             Flag
	Voice                  Flag
	Transform              Flag
	FragmentReceived       Flag
	NetworkVarSync         Flag
	ChatMessage            Flag
}

var Response = ResponseStruct{
	IDAssign:               0x03,
	PlayerList:             0x05,
	LobbyList:              0x06,
	Error:                  0xFF,
	ClosingCon:             0xF0,
	LobbyListChanged:       0x0A,
	LobbyClosing:           0xF1,
	PlayerTransforms:       0xB0,
	LobbyInfo:              0x09,
	WorldState:             0xC2,
	ItemIntInf:             0xC5,
	ItemList:               0xA1,
	ItemPickup:             0xC6,
	ItemDrop:               0xC7,
	InventorySwitch:        0xC8,
	StartMap:               0xC9,
	InteractableMessage:    0x0D,
	CodeInteractionMessage: 0x0E,
	PlayerData:             0xC4, // warning: contents explosive
	Voice:                  0x0C,
	Transform:              0x0F,
	FragmentReceived:       0xDF,
	NetworkVarSync:         0xEE,
	ChatMessage:            0xE0,
}
