package exocomMock

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type ExoCom struct {
	sync.Mutex
	ServerPort       int
	Services         map[string]*websocket.Conn
	ReceivedMessages []Message
}

type Message struct {
	Name         string      `json:"name"`
	Sender       string      `json:"sender"`
	Payload      interface{} `json:"payload"`
	ResponseTo   string      `json:"responseTo"`
	ID           string      `json:"id"`
	Timestamp    int         `json:"timestamp"`
	ResponseTime int         `json:"timestamp"`
}

func New() *ExoCom {
	log.Println("EXOCOM: ExoCom initialized!")
	return &ExoCom{
		ServerPort:       0,
		Services:         make(map[string]*websocket.Conn),
		ReceivedMessages: make([]Message, 0),
	}
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
		var incoming Message
		err := websocket.JSON.Receive(ws, &incoming)
		if err != nil {
			log.Fatal(err)
		}
		exocom.Lock()
		if incoming.Name == "exocom.register-service" {
			exocom.RegisterService(incoming.Sender, ws)
		}
		fmt.Printf("EXOCOM: ---- INCOMING MESSAGE ---- : %#v\n", incoming)
		exocom.ReceivedMessages = append(exocom.ReceivedMessages, incoming)
		exocom.Unlock()
	}

	http.Handle("/services", websocket.Handler(onMessage))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("EXOCOM: Listener is done")
}
