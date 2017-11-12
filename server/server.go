package server

import (
	encoding "encoding/json"
	"io"
	"log"
	"net"
	"sync"

	"github.com/mostafa-asg/hodhod/event"
	"github.com/mostafa-asg/hodhod/util"
)

type user struct {
	con      net.Conn
	nickname string
}

//Config configures the server
type Config struct {
	Binding string //syntax is host:port
}

//Server represent the serer
type Server struct {
	Config     *Config
	HasStarted chan bool
	listener   net.Listener

	chatrooms      map[string]map[string]*user
	chatroomsMutex sync.Mutex
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

	decoder := encoding.NewDecoder(con)

	var metadata event.Metadata
	var joinEvent event.Join

	for {
		err := decoder.Decode(&metadata)
		if err != nil {
			if err == io.EOF {
				return
			}
			log.Println("error decodign metadata request from client", err)
			return
		}

		switch metadata.EventType {
		case "join":
			err := decoder.Decode(&joinEvent)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("error decodign join request from client", err)
				return
			}

			users := s.getChatroomUsers(joinEvent.Chatroom)
			uuid := s.addUserToChatroom(joinEvent.Chatroom, &user{con: con, nickname: joinEvent.Nickname})

			encoder := encoding.NewEncoder(con)
			encoder.Encode(&event.ChatroomUsers{Users: users})

			//TODO remove this line
			log.Println(joinEvent.Nickname + "->" + uuid)
		}
	}
}

func (s *Server) addUserToChatroom(chatroomName string, userInfo *user) string {

	uuid, _ := util.NewUUID()

	s.chatroomsMutex.Lock()
	users, ok := s.chatrooms[chatroomName]
	if !ok {
		// This user is the first user that has joined the chatroom
		users = make(map[string]*user)
		s.chatrooms[chatroomName] = users
	}
	users[uuid] = userInfo
	s.chatroomsMutex.Unlock()

	return uuid
}

func (s *Server) getChatroomUsers(chatroomName string) map[string]string {
	s.chatroomsMutex.Lock()
	users, ok := s.chatrooms[chatroomName]
	s.chatroomsMutex.Unlock()

	if !ok {
		return make(map[string]string)
	}

	result := make(map[string]string)
	for uuid, userInfo := range users {
		result[uuid] = userInfo.nickname
	}

	return result
}
