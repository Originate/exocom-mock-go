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
	port             int
	services         map[string]*websocket.Conn
	ReceivedMessages []Message
	registerCh       chan *websocket.Conn
	doneCh           chan bool
	errCh            chan error
}

type Message struct {
	Name         string      `json:"name,omitempty"`
	Sender       string      `json:"sender,omitempty"`
	Payload      interface{} `json:"payload,omitempty"`
	ResponseTo   string      `json:"responseTo,omitempty"`
	ID           string      `json:"id,omitempty"`
	Timestamp    int         `json:"timestamp,omitempty"`
	ResponseTime int         `json:"timestamp,omitempty"`
}

func New() *ExoCom {
	log.Println("EXOCOM: ExoCom initialized!")
	return &ExoCom{
		port:             0,
		services:         make(map[string]*websocket.Conn),
		ReceivedMessages: make([]Message, 0),
	}
}

func (exocom *ExoCom) RegisterService(name string, ws *websocket.Conn) {
	exocom.services[name] = ws
}

func (exocom *ExoCom) Close() {
	exocom.doneCh <- true
}

func (exocom *ExoCom) listenToMessages(ws *websocket.Conn) {
	for {
		select {
		case <-exocom.doneCh:
			return
		default:
			var incoming Message
			if err := websocket.JSON.Receive(ws, &incoming); err != nil {
				log.Fatal(err)
			}
			exocom.Lock()
			exocom.saveMessage(incoming)
			exocom.Unlock()
			return
		}
	}
}

func (exocom *ExoCom) saveMessage(message Message) {
	exocom.ReceivedMessages = append(exocom.ReceivedMessages, message)
}

func (exocom *ExoCom) Listen(port int) {
	exocom.port = port

	onConnection := func(ws *websocket.Conn) {
		var incoming Message
		if err := websocket.JSON.Receive(ws, &incoming); err != nil {
			log.Fatal(err)
		}
		if incoming.Name == "exocom.register-service" {
			exocom.RegisterService(incoming.Sender, ws)
			exocom.saveMessage(incoming)
			exocom.listenToMessages(ws)
		}
	}

	http.Handle("/services", websocket.Handler(onConnection))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("EXOCOM: Listener is done")
}
