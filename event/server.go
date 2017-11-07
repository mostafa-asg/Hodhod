package event

// NewUserJoined is fired by server whenever a user joined to the chatroom
type NewUserJoined struct {
	Nickname string //User that has joined
}
