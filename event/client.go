package event

// Join is used by users to join to a specific chat room
type Join struct {
	Nickname string // Name that be visible to other users on the specific chat room
	Chatroom string // chatroom name
}

// Leave is used by users to leave a chatroom
type Leave struct {
	Chatroom string //Chatroom name
}

//Message is used by users to send private message to other users
type Message struct {
	RecieverID string
	Message    string
}

// Broadcast broadcast a meessage to all users on a specific chatroom
type Broadcast struct {
	Chatroom string //Chatroom name
	Message  string
}
