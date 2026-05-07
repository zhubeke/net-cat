package server

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	maxConnections = 10
)

var banner = `
Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    '.       | ` + "`" + `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     '-\       '--'
[ENTER YOUR NAME]: `


type Message struct {
	text   string
	sender *Client 
}

type Server struct {
	clients     map[*Client]bool
	mu          sync.Mutex
	messages    []string
	messageChan chan Message
	joinChan    chan *Client
	leaveChan   chan *Client
}

func NewServer() *Server {
	return &Server{
		clients:     make(map[*Client]bool),
		messages:    make([]string, 0),
		messageChan: make(chan Message, 16), 
		joinChan:    make(chan *Client),
		leaveChan:   make(chan *Client),
	}
}

func (s *Server) Start(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}
	defer listener.Close()

	fmt.Printf("Listening on the port :%s\n", port)

	go s.broadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		s.mu.Lock()
		if len(s.clients) >= maxConnections {
			s.mu.Unlock()
			conn.Write([]byte("Server is full. Try again later.\n"))
			conn.Close()
			continue
		}
		s.mu.Unlock()

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	conn.Write([]byte(banner))

	reader := bufio.NewReader(conn)
	name, err := reader.ReadString('\n')
	if err != nil {
		conn.Close()
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		conn.Write([]byte("Name cannot be empty. Disconnecting...\n"))
		conn.Close()
		return
	}

	s.mu.Lock()
	for client := range s.clients {
		if client.GetName() == name {
			s.mu.Unlock()
			conn.Write([]byte("Name already taken. Please reconnect with a different name.\n"))
			conn.Close()
			return
		}
	}
	s.mu.Unlock()

	client := NewClient(conn)
	client.SetName(name)

	
	s.mu.Lock()
	for _, msg := range s.messages {
		client.SendMessage(msg + "\n")
	}
	s.mu.Unlock()

	s.joinChan <- client

	s.handleClientMessages(client, reader)
}

func (s *Server) handleClientMessages(client *Client, reader *bufio.Reader) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		message := strings.TrimSpace(scanner.Text())
		if message == "" {
			client.sendPrompt()
			continue
		}

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		formattedMsg := fmt.Sprintf("[%s][%s]:%s", timestamp, client.GetName(), message)

		s.mu.Lock()
		s.messages = append(s.messages, formattedMsg)
		s.mu.Unlock()

		s.messageChan <- Message{text: formattedMsg, sender: client}
	}

	s.leaveChan <- client
}

func (s *Server) broadcast() {
	for {
		select {
		case client := <-s.joinChan:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()

			joinMsg := fmt.Sprintf("%s has joined our chat...", client.GetName())
			s.mu.Lock()
			s.messages = append(s.messages, joinMsg)
			s.mu.Unlock()

			go func(msg string, joiner *Client) {
				s.messageChan <- Message{text: msg, sender: joiner}
			}(joinMsg, client)

		case client := <-s.leaveChan:
			s.mu.Lock()
			if _, exists := s.clients[client]; exists {
				delete(s.clients, client)
			}
			s.mu.Unlock()

			leaveMsg := fmt.Sprintf("%s has left our chat...", client.GetName())
			s.mu.Lock()
			s.messages = append(s.messages, leaveMsg)
			s.mu.Unlock()

			client.conn.Close()

			go func(msg string) {
				s.messageChan <- Message{text: msg, sender: nil}
			}(leaveMsg)

		case msg := <-s.messageChan:
			s.mu.Lock()
			for client := range s.clients {
				if msg.sender == client {
					client.sendPrompt()
					continue
				}
				if err := client.SendMessage("\n" + msg.text + "\n"); err != nil {
					delete(s.clients, client)
					client.conn.Close()
					continue
				}
				client.sendPrompt()
			}
			s.mu.Unlock()
		}
	}
}