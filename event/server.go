package event

import (
	"github.com/mostafa-asg/hodhod/model"
)

// NewUserJoined is fired by server whenever a user joined to the chatroom
type NewUserJoined struct {
	Nickname string //User that has joined
}

// NewMessage sends message to specific user
type NewMessage struct {
	FromID  string
	Message string
}

// NewBroadcastMessage represent the broadcast message
type NewBroadcastMessage struct {
	FromID  string
	Message string
}

// JoinResponse is used by server whenever a user joined
type JoinResponse struct {
	Users  []*model.User // all availble users on this chatroom
	YourID string        // your id
}
