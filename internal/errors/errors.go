package errors

type Error struct {
	Message     string
	Description string
}

// Sent when the server gets corrupted or invalid packet payload
func InvalidPacket(description string) Error {
	return Error{"INVALID_PACKET", description}
}

// Sent when expected packet does not gets received
func ExpectedPacket(description string) Error {
	return Error{"NO_EXPECTED_PACKET", description}
}

// Sent when important data does not match between the cliend and the server
func DataMismatch(description string) Error {
	return Error{"DATA_MISMATCH", description}
}

// Sent when data provided by the client is incorrect
func InvalidData(description string) Error {
	return Error{"INVALID_DATA", description}
}

// Sent when client performs an unauthorized action
func Unauthorized(description string) Error {
	return Error{"UNAUTHORIZED", description}
}
