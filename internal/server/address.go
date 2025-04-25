package server

import "strconv"

type Address struct {
	IP   string
	Port int
}

// Returns the address in the following string form : 127.0.0.1:8080
func (addr *Address) String() string {
	return addr.IP + ":" + strconv.FormatInt((int64)(addr.Port), 10)
}
