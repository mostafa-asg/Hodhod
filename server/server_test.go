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

type clientInfo struct {
	connection net.Conn
	id         string
}

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
	nickname string, chatroom string, availableUsers []string, futureJoinUsers []string, wg *sync.WaitGroup) *clientInfo {
	con, err := net.Dial("tcp4", serverAddr)
	if err != nil {
		t.Fatal("Could not connect to server", err)
	}

	encoder := encoding.NewEncoder(con)
	decoder := encoding.NewDecoder(con)

	encoder.Encode(&model.Metadata{EventType: "join"})
	encoder.Encode(&event.Join{Nickname: nickname, Chatroom: chatroom})

	var joinResponse event.JoinResponse
	err = decoder.Decode(&joinResponse)
	if err != nil {
		t.Fatal("error in decoding chatroom users", err)
	}

	actualUsers := len(joinResponse.Users)
	if len(availableUsers) != actualUsers {
		t.Errorf("Expected %d user(s) in chatroom but find %d user(s)", len(availableUsers), actualUsers)
	}

	if actualUsers > 0 {
		for _, user := range joinResponse.Users {
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

	return &clientInfo{connection: con, id: joinResponse.YourID}
}

func contains(slice []string, data string) bool {
	for _, val := range slice {
		if val == data {
			return true
		}
	}

	return false
}

func TestJoiningUsersToChatrooms(t *testing.T) {

	s := startTheServer(t)

	var wg sync.WaitGroup

	client1 := connectToServer(t, s.HostAndPort(), "John", "room1", nil, []string{"Sara", "Bill", "Kevin"}, &wg)

	client2 := connectToServer(t, s.HostAndPort(), "Sara", "room1", []string{"John"}, []string{"Bill", "Kevin"}, &wg)

	client3 := connectToServer(t, s.HostAndPort(), "Bill", "room1", []string{"John", "Sara"}, []string{"Kevin"}, &wg)

	client4 := connectToServer(t, s.HostAndPort(), "Kevin", "room1", []string{"John", "Sara", "Bill"}, nil, &wg)

	wg.Wait()

	//Stop clients
	client1.connection.Close()
	client2.connection.Close()
	client3.connection.Close()
	client4.connection.Close()

	//stop the server
	err := s.Stop()
	if err != nil {
		t.Error("Error in closing the server", err)
	}
}

func sendMessage(sender *clientInfo, receiver *clientInfo, chatroom string, message string) {

	encoder := encoding.NewEncoder(sender.connection)
	encoder.Encode(&model.Metadata{EventType: "send_msg"})
	encoder.Encode(&event.Message{Chatroom: chatroom, FromID: sender.id, RecieverID: receiver.id, Message: message})

}

func receiveMessage(t *testing.T, client *clientInfo, expectedMessages []string) {
	decoder := encoding.NewDecoder(client.connection)

	var metadata model.Metadata
	var msg event.NewMessage
	count := 0

	for count < len(expectedMessages) {
		err := decoder.Decode(&metadata)
		if err != nil {
			t.Fatal("error decoding Metadata")
		}

		if metadata.EventType == "new_msg" {
			err := decoder.Decode(&msg)
			if err != nil {
				t.Error("error decoding Metadata")
			}
			if !contains(expectedMessages, msg.Message) {
				t.Errorf("invalid message : %s", msg.Message)
			}
			count++
		}
	}
}

func broadcastMessage(sender *clientInfo, chatroom string, message string) {
	encoder := encoding.NewEncoder(sender.connection)
	encoder.Encode(&model.Metadata{EventType: "broadcast_msg"})
	encoder.Encode(&event.Broadcast{Chatroom: chatroom, FromID: sender.id, Message: message})
}

func receiveBroadcastMessage(t *testing.T, client *clientInfo, expectedMessages []string, wg *sync.WaitGroup) {
	decoder := encoding.NewDecoder(client.connection)

	var metadata model.Metadata
	var msg event.NewBroadcastMessage
	count := 0

	for count < len(expectedMessages) {
		err := decoder.Decode(&metadata)
		if err != nil {
			t.Fatal("error decoding Metadata")
		}

		if metadata.EventType == "new_broadcast_msg" {
			err := decoder.Decode(&msg)
			if err != nil {
				t.Fatal("error decoding NewBroadcastMessage")
			}
			if !contains(expectedMessages, msg.Message) {
				t.Errorf("invalid message : %s", msg.Message)
			}
			count++
		}
	}
	wg.Done()
}

func TestMessageBetweenTwoUser(t *testing.T) {
	s := startTheServer(t)
	var wg sync.WaitGroup

	client1 := connectToServer(t, s.HostAndPort(), "Ben", "room1", nil, []string{"Tom"}, &wg)

	client2 := connectToServer(t, s.HostAndPort(), "Tom", "room1", []string{"Ben"}, nil, &wg)

	wg.Wait()

	sendMessage(client1, client2, "room1", "Hi")
	sendMessage(client1, client2, "room1", "How are you Tom?")

	receiveMessage(t, client2, []string{"Hi", "How are you Tom?"})

	client1.connection.Close()
	client2.connection.Close()

	//stop the server
	err := s.Stop()
	if err != nil {
		t.Error("Error in closing the server", err)
	}
}

func TestBroadcastMessage(t *testing.T) {

	s := startTheServer(t)
	var wg sync.WaitGroup

	client1 := connectToServer(t, s.HostAndPort(), "John", "room1", nil, []string{"Sara", "Bill", "Kevin"}, &wg)

	client2 := connectToServer(t, s.HostAndPort(), "Sara", "room1", []string{"John"}, []string{"Bill", "Kevin"}, &wg)

	client3 := connectToServer(t, s.HostAndPort(), "Bill", "room1", []string{"John", "Sara"}, []string{"Kevin"}, &wg)

	client4 := connectToServer(t, s.HostAndPort(), "Kevin", "room1", []string{"John", "Sara", "Bill"}, nil, &wg)

	wg.Wait()

	broadcastMessage(client1, "room1", "Hi guys")
	broadcastMessage(client1, "room1", "What's up?")

	wg.Add(3)
	go receiveBroadcastMessage(t, client2, []string{"Hi guys", "What's up?"}, &wg)
	go receiveBroadcastMessage(t, client3, []string{"Hi guys", "What's up?"}, &wg)
	go receiveBroadcastMessage(t, client4, []string{"Hi guys", "What's up?"}, &wg)
	wg.Wait()

	//Stop clients
	client1.connection.Close()
	client2.connection.Close()
	client3.connection.Close()
	client4.connection.Close()

	//stop the server
	err := s.Stop()
	if err != nil {
		t.Error("Error in closing the server", err)
	}
}
