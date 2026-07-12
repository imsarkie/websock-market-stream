package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Server struct{

}

func NewServer() (s *Server){
	return &Server{}
}

func (s *Server) home(w http.ResponseWriter, r *http.Request){
	w.Write([]byte ("Market Stream Server Running"))
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
	}

	defer conn.Close()
	log.Println("Browser Connected!")

	

	for{
		err = conn.WriteMessage(
			websocket.TextMessage,
			[]byte ("Hello from server!"),
		) 
		if err != nil {
			log.Println(err)
			break
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}
		log.Printf("Browser says: %s\n", msg)
	}
}

func(s *Server) Start() error{
	http.HandleFunc("/", s.home)
	http.HandleFunc("/ws", s.handleWS)

	return http.ListenAndServe(":8080", nil)
}