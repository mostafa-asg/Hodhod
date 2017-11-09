package server_test

import (
	"testing"

	"github.com/mostafa-asg/hodhod/server"
)

func TestStartAndStopTheServer(t *testing.T) {

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

	err := s.Stop()
	if err != nil {
		t.Error("Error in closing the server", err)
	}
}
