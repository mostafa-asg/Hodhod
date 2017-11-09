package server_test

import (
	"testing"

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

func TestJoiningUsersToChatrooms(t *testing.T) {

}
