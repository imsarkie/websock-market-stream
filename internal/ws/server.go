package ws

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/imsarkie/websock-market-stream/internal/history"
)

type Server struct{
	clients		map[*websocket.Conn]bool

	history 	*history.Store
}

func NewServer(history *history.Store) (s *Server){
	return &Server{
		clients: make(map[*websocket.Conn]bool),
		history: history,
	}
}

func(s *Server) Start() error{
	fs := http.FileServer(http.Dir("./web"))
	// http.HandleFunc("/", s.home)
	http.Handle("/", fs)
	http.HandleFunc("/ws", s.handleWS)
	http.HandleFunc("/history", s.handleHistory)

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

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request){

	w.Header().Set("content-type", "application/json")

	candles := s.history.GetAll()
	
	err := json.NewEncoder(w).Encode(candles)
	if err != nil {
		http.Error(
			w,
			"Failed to Encode.",
			http.StatusInternalServerError,
		)
	}
}