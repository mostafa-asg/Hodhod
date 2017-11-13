package server

import (
	encoding "encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sync"

	"github.com/mostafa-asg/hodhod/event"
	"github.com/mostafa-asg/hodhod/model"
	"github.com/mostafa-asg/hodhod/util"
)

//Config configures the server
type Config struct {
	Binding string //syntax is host:port
}

//Server represent the serer
type Server struct {
	Config     *Config
	HasStarted chan bool
	listener   net.Listener

	chatrooms      map[string][]*model.User
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
	s.chatrooms = make(map[string][]*model.User)

	return s
}

func accept(s *Server, con net.Conn) {

	decoder := encoding.NewDecoder(con)

	var metadata model.Metadata

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
			var joinEvent event.Join
			err := decoder.Decode(&joinEvent)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("error decoding join request from client", err)
				return
			}

			users := s.getChatroomUsers(joinEvent.Chatroom)
			uuid := s.addUserToChatroom(joinEvent.Chatroom, &model.User{Connection: con, Nickname: joinEvent.Nickname})

			encoder := encoding.NewEncoder(con)
			//Send available users to the newly joined user
			encoder.Encode(&event.JoinResponse{Users: users, YourID: uuid})

			//TODO remove this line
			log.Println(joinEvent.Nickname + "->" + uuid)

			//Notify to other users that someone has joined
			go s.notiftyNewUserJoined(users, joinEvent.Nickname)
		case "send_msg":
			var msg event.Message
			err := decoder.Decode(&msg)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Println("error decoding message request from client", err)
				return
			}
			users := s.chatrooms[msg.Chatroom]
			recipient, err := find(users, msg.RecieverID)
			if err != nil {
				//TODO send error to the client
				break
			}
			encoder := encoding.NewEncoder(recipient.Connection)
			encoder.Encode(&model.Metadata{EventType: "new_msg"})
			encoder.Encode(&event.NewMessage{FromID: msg.FromID, Message: msg.Message})
		}
	}
}

func find(users []*model.User, userID string) (usr *model.User, err error) {
	for _, user := range users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, errors.New("User not found")
}

func (s *Server) notiftyNewUserJoined(others []*model.User, newUserNickname string) {

	for _, user := range others {
		go func(u *model.User) {

			encoder := encoding.NewEncoder(u.Connection)
			encoder.Encode(&model.Metadata{EventType: "newUser"})
			encoder.Encode(&event.NewUserJoined{Nickname: newUserNickname})

		}(user)
	}
}

func (s *Server) addUserToChatroom(chatroomName string, user *model.User) string {

	uuid, _ := util.NewUUID()

	s.chatroomsMutex.Lock()
	users, ok := s.chatrooms[chatroomName]
	if !ok {
		users = make([]*model.User, 0)
	}

	user.ID = uuid
	users = append(users, user)
	s.chatrooms[chatroomName] = users
	s.chatroomsMutex.Unlock()

	return uuid
}

func (s *Server) getChatroomUsers(chatroomName string) []*model.User {
	s.chatroomsMutex.Lock()
	users, ok := s.chatrooms[chatroomName]
	s.chatroomsMutex.Unlock()

	if !ok {
		return make([]*model.User, 0)
	}

	return users
}
