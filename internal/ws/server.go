package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct{
	clients		map[*websocket.Conn]bool
}

func NewServer() (s *Server){
	return &Server{
		clients: make(map[*websocket.Conn]bool),
	}
}

func(s *Server) Start() error{
	http.HandleFunc("/", s.home)
	http.HandleFunc("/ws", s.handleWS)

	return http.ListenAndServe(":8080", nil)
}

func (s *Server) home(w http.ResponseWriter, r *http.Request){
	// w.Write([]byte ("Market Stream Server Running"))
	http.ServeFile(w, r, "web/index.html")
}

// Create an upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool{
		return true
	},
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request){
	
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	defer func ()  {
		conn.Close()
		delete(s.clients, conn)

		log.Printf("Client disconnected")
	} ()

	log.Println("Browser Connected!")

	s.clients[conn] = true
	log.Printf("Client connected. Total clients: %d\n", len(s.clients))

	for{
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("Browser says: %s\n", msg)

		err = conn.WriteMessage(
			websocket.TextMessage,
			[]byte ("Hello from server!"),
		) 
		if err != nil {
			log.Println(err)
			break
		}

	}
}

// Method Broadcast()
func (s *Server) Broadcast(v any) error{

	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	for client := range s.clients{
		
		err := client.WriteMessage(
			websocket.TextMessage,
			data,
		)
		if err != nil {
			log.Println(err)
			client.Close()
			delete(s.clients, client)
		}
	}
	return nil
}
