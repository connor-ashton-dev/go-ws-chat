package server

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
	Port  int
	mu    sync.RWMutex
}

func New(port int) *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
		Port:  port,
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	fmt.Println("New incomming connection from client:", ws.RemoteAddr())

	// maps are not concurrency safe
	s.mu.Lock()
	s.conns[ws] = true
	s.mu.Unlock()

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buf := make([]byte, 1024)
	for {
		n, err := ws.Read(buf)
		if err != nil {
			// connection has closed itself
			if err == io.EOF {
				return
			}
			println("Read error: ", err)
			continue
		}

		msg := buf[:n]
		s.broadcastMessage(ws, msg)
	}
}

func (s *Server) broadcastMessage(sender *websocket.Conn, msg []byte) {
	s.mu.RLock()
	var wg sync.WaitGroup
	inactiveCons := make([]*websocket.Conn, 0)

	for conn := range s.conns {
		if conn != sender {
			wg.Add(1)
			go func(conn *websocket.Conn) {
				defer wg.Done()
				_, err := conn.Write([]byte(string(msg)))
				if err != nil {
					println("Write error during broadcast:", err)
					inactiveCons = append(inactiveCons, conn)
				}
			}(conn)
		}
	}

	s.mu.RUnlock()
	wg.Wait()

	s.mu.Lock()
	for _, conn := range inactiveCons {
		if _, ok := s.conns[conn]; ok {
			delete(s.conns, conn)
			fmt.Println("Deleted inactive client: ", conn.RemoteAddr())
		}
	}
	s.mu.Unlock()
}

func (server *Server) Start() {
	http.Handle("/ws", websocket.Handler(server.handleWS))
	serverPort := fmt.Sprintf(":%d", server.Port)
	if err := http.ListenAndServe(serverPort, nil); err != nil {
		log.Fatal(err)
	}
}
