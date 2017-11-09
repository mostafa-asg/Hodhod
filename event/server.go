package event

// NewUserJoined is fired by server whenever a user joined to the chatroom
type NewUserJoined struct {
	Nickname string //User that has joined
}

// NewMessage sends message to specific user
type NewMessage struct {
	Message string
}

// ChatroomUsers is used by server whenever a user joined to show all availble user
type ChatroomUsers struct {
	Users map[string]string
}
