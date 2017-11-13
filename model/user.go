package model

import (
	"net"
)

// User information
type User struct {
	Connection net.Conn `json:"-"` //tcp connection
	Nickname   string
	ID         string // user unique id
}
