package server_test

import (
	encoding "encoding/json"
	"net"
	"sync"
	"testing"

	"github.com/mostafa-asg/hodhod/event"
	"github.com/mostafa-asg/hodhod/model"
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
	nickname string, chatroom string, availableUsers []string, futureJoinUsers []string, wg *sync.WaitGroup) net.Conn {
	con, err := net.Dial("tcp4", serverAddr)
	if err != nil {
		t.Fatal("Could not connect to server", err)
	}

	encoder := encoding.NewEncoder(con)
	decoder := encoding.NewDecoder(con)

	encoder.Encode(&model.Metadata{EventType: "join"})
	encoder.Encode(&event.Join{Nickname: nickname, Chatroom: chatroom})

	var chatroomUsers event.ChatroomUsers
	err = decoder.Decode(&chatroomUsers)
	if err != nil {
		t.Fatal("error in decoding chatroom users", err)
	}

	actualUsers := len(chatroomUsers.Users)
	if len(availableUsers) != actualUsers {
		t.Errorf("Expected %d user(s) in chatroom but find %d user(s)", len(availableUsers), actualUsers)
	}

	if actualUsers > 0 {
		for _, user := range chatroomUsers.Users {
			if !contains(availableUsers, user.Nickname) {
				t.Errorf("user %s not found in chatroom", user.Nickname)
			}
		}
	}

	wg.Add(1)
	go func() {
		var metadata model.Metadata
		var join event.NewUserJoined

		userJoinedCount := 0

		for userJoinedCount < len(futureJoinUsers) {

			err = decoder.Decode(&metadata)
			if err != nil {
				t.Error("cannot decode Metadata", err)
			}

			if metadata.EventType == "newUser" {
				err := decoder.Decode(&join)
				if err != nil {
					t.Error("cannot decode NewUserJoined", err)
				}
				if !contains(futureJoinUsers, join.Nickname) {
					t.Errorf("%s is not in future join list", join.Nickname)
				} else {
					userJoinedCount++
				}
			}
		}

		wg.Done()
	}()

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

	var wg sync.WaitGroup

	client1 := connectToServer(t, s.HostAndPort(), "John", "room1", nil, []string{"Sara", "Bill", "Kevin"}, &wg)
	defer client1.Close()

	client2 := connectToServer(t, s.HostAndPort(), "Sara", "room1", []string{"John"}, []string{"Bill", "Kevin"}, &wg)
	defer client2.Close()

	client3 := connectToServer(t, s.HostAndPort(), "Bill", "room1", []string{"John", "Sara"}, []string{"Kevin"}, &wg)
	defer client3.Close()

	client4 := connectToServer(t, s.HostAndPort(), "Kevin", "room1", []string{"John", "Sara", "Bill"}, nil, &wg)
	defer client4.Close()

	wg.Wait()
}
