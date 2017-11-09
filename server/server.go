package server

import (
	"log"
	"net"
)

type user struct {
	con      net.Conn
	nickname string
}

//Config configures the server
type Config struct {
	Binding string //syntax is host:port
}

type Server struct {
	Config     *Config
	HasStarted chan bool
	listener   net.Listener
	chatrooms  map[string]map[string]*user
}

//HostAndPort return host and port of the server
func (s *Server) HostAndPort() string {
	return s.listener.Addr().String()
}

// Start the server and accept incomming connections
// This method will be block the caller
func (s *Server) Start() error {
	l, err := net.Listen("tcp4", s.Config.Binding)
	if err != nil {
		return err
	}
	s.listener = l
	log.Printf("Server is listening on : %s\n", s.HostAndPort())
	// Tell the others that server has started
	s.HasStarted <- true

	for {
		con, err := l.Accept()
		if err != nil {
			log.Println("error while accepting tcp connection", err)
			continue
		}

		go accept(s, con)
	}
}

// Stop the server
func (s *Server) Stop() error {
	return s.listener.Close()
}

// New Create new server
func New(conf *Config) *Server {

	if conf.Binding == "" {
		conf.Binding = ":7474"
	}

	s := &Server{Config: conf}
	s.HasStarted = make(chan bool, 1)
	s.chatrooms = make(map[string]map[string]*user)

	return s
}

func accept(s *Server, con net.Conn) {

	defer con.Close()

}
