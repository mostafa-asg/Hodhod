package server_test

import (
	"encoding/gob"
	"net"
	"testing"

	"github.com/mostafa-asg/hodhod/event"
	"github.com/mostafa-asg/hodhod/server"
)

func startTheServer(t *testing.T) *server.Server {
	opts := &server.Config{
		Binding: "localhost:0",
	}

	s := server.New(opts)
	go func() {
		t.Log("Starting the server ...")
		err := s.Start()
		if err != nil {
			t.Fatal("Server could not start", err)
		}
	}()

	//Wait until server started
	<-s.HasStarted
	return s
}

func TestStartAndStopTheServer(t *testing.T) {

	s := startTheServer(t)

	err := s.Stop()
	if err != nil {
		t.Error("Error in closing the server", err)
	}
}

func connectToServer(t *testing.T, serverAddr string,
	nickname string, chatroom string, expectedUsers int, expectedNames []string) net.Conn {
	con, err := net.Dial("tcp4", serverAddr)
	if err != nil {
		t.Fatal("Could not connect to server", err)
	}

	encoder := gob.NewEncoder(con)
	decoder := gob.NewDecoder(con)

	encoder.Encode(&event.Metadata{EventType: "join"})
	encoder.Encode(&event.Join{Nickname: nickname, Chatroom: chatroom})

	var chatroomUsers event.ChatroomUsers
	err = decoder.Decode(&chatroomUsers)
	if err != nil {
		t.Fatal("error in decoding chatroom users", err)
	}

	actualUsers := len(chatroomUsers.Users)
	if expectedUsers != actualUsers {
		t.Errorf("Expected %d user(s) in chatroom but find %d user(s)", expectedUsers, actualUsers)
	}

	if actualUsers > 0 {
		for _, userNickname := range chatroomUsers.Users {
			if !contains(expectedNames, userNickname) {
				t.Errorf("user %s not found in chatroom", userNickname)
			}
		}
	}

	return con
}

func contains(users []string, user string) bool {
	for _, val := range users {
		if val == user {
			return true
		}
	}

	return false
}

func TestJoiningUsersToChatrooms(t *testing.T) {

	s := startTheServer(t)

	client1 := connectToServer(t, s.HostAndPort(), "John", "room1", 0, nil)
	defer client1.Close()

	client2 := connectToServer(t, s.HostAndPort(), "Sara", "room1", 1, []string{"John"})
	defer client2.Close()

	client3 := connectToServer(t, s.HostAndPort(), "Bill", "room1", 2, []string{"John", "Sara"})
	defer client3.Close()

	client4 := connectToServer(t, s.HostAndPort(), "Kevin", "room1", 3, []string{"John", "Sara", "Bill"})
	defer client4.Close()
}
