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

type Server struct {
	clients      map[*Client]bool
	mu           sync.Mutex
	messages     []string
	messageChan  chan string
	joinChan     chan *Client
	leaveChan    chan *Client
}

func NewServer() *Server {
	return &Server{
		clients:     make(map[*Client]bool),
		messages:    make([]string, 0),
		messageChan: make(chan string),
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

	// Start broadcast goroutine
	go s.broadcast()

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}

		// Check connection limit
		s.mu.Lock()
		if len(s.clients) >= maxConnections {
			s.mu.Unlock()
			conn.Write([]byte("Server is full. Try again later.\n"))
			conn.Close()
			continue
		}
		s.mu.Unlock()

		// Handle new client
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	// Send welcome banner
	conn.Write([]byte(banner))

	// Get client name
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

	// Check for duplicate names
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

	// Create new client
	client := NewClient(conn)
	client.SetName(name)

	// Send previous messages to new client
	s.mu.Lock()
	for _, msg := range s.messages {
		client.SendMessage(msg + "\n")
	}
	s.mu.Unlock()

	// Add client to server
	s.joinChan <- client

	// Handle client messages
	go s.handleClientMessages(client, reader)
}

func (s *Server) handleClientMessages(client *Client, reader *bufio.Reader) {
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			// Client disconnected
			s.leaveChan <- client
			return
		}

		message = strings.TrimSpace(message)
		if message == "" {
			continue // Don't broadcast empty messages
		}

		// Format message with timestamp and name
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		formattedMsg := fmt.Sprintf("[%s][%s]:%s", timestamp, client.GetName(), message)
		
		// Store message history
		s.mu.Lock()
		s.messages = append(s.messages, formattedMsg)
		s.mu.Unlock()

		// Broadcast message
		s.messageChan <- formattedMsg
	}
}

func (s *Server) broadcast() {
	for {
		select {
		case client := <-s.joinChan:
			s.mu.Lock()
			s.clients[client] = true
			s.mu.Unlock()

			// Notify other clients
			joinMsg := fmt.Sprintf("%s has joined our chat...", client.GetName())
			s.messageChan <- joinMsg
			
			// Add to history
			s.mu.Lock()
			s.messages = append(s.messages, joinMsg)
			s.mu.Unlock()

		case client := <-s.leaveChan:
			s.mu.Lock()
			if _, exists := s.clients[client]; exists {
				delete(s.clients, client)
			}
			s.mu.Unlock()

			// Notify other clients
			leaveMsg := fmt.Sprintf("%s has left our chat...", client.GetName())
			s.messageChan <- leaveMsg
			
			// Add to history
			s.mu.Lock()
			s.messages = append(s.messages, leaveMsg)
			s.mu.Unlock()
			
			client.conn.Close()

		case message := <-s.messageChan:
			// Broadcast message to all clients
			s.mu.Lock()
			for client := range s.clients {
				client.SendMessage("\n" + message + "\n")
			}
			s.mu.Unlock()
		}
	}
}
