package exocomMock

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

type ExoCom struct {
	ServerPort int
	Services   map[string]*websocket.Conn
}

type Message struct {
	Sender       string `json:"sender"`
	Name         string `json:"name"`
	Payload      string `json:"payload"`
	ResponseTo   string `json:"responseTo"`
	ID           string `json:"id"`
	Timestamp    int    `json:"timestamp"`
	ResponseTime int    `json:"timestamp"`
}

func New() *ExoCom {
	log.Println("EXOCOM: ExoCom initialized!")
	return &ExoCom{0, make(map[string]*websocket.Conn)}
}

func (exocom *ExoCom) RegisterService(name string, ws *websocket.Conn) {
	exocom.Services[name] = ws
}

func (exocom *ExoCom) Close() {

}

func (exocom *ExoCom) Listen(port int) {
	log.Println("EXOCOM: Starting listener.")
	exocom.ServerPort = port
	onMessage := func(ws *websocket.Conn) {
		var incoming *Message
		err := websocket.JSON.Receive(ws, &incoming)
		if err != nil {
			log.Fatal(err)
		}
		if incoming.Name == "exocom.register-service" {
			exocom.RegisterService(incoming.Sender, ws)
		}
	}

	http.Handle("/services", websocket.Handler(onMessage))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("EXOCOM: Listener is done")
}
