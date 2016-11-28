package exocomMock

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/fatih/color"

	"golang.org/x/net/websocket"
)

var red = color.New(color.FgRed).SprintFunc()
var cyan = color.New(color.FgHiCyan).SprintFunc()

type ExoCom struct {
	sync.Mutex
	port             int
	services         map[string]*websocket.Conn
	ReceivedMessages []Message
	doneCh           chan bool
	messageCh        chan Message
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
		doneCh:           make(chan bool, 1),
		messageCh:        make(chan Message),
		errCh:            make(chan error),
	}
}

func (exocom *ExoCom) RegisterService(name string, ws *websocket.Conn) {
	exocom.services[name] = ws
}

func (exocom *ExoCom) Close() {
	log.Println(cyan("EXOCOM: Sending true to doneCh"))
	exocom.doneCh <- true
	log.Println(cyan("EXOCOM: Sending true to doneCh sent"))
}

func (exocom *ExoCom) listenToMessages(ws *websocket.Conn) {
	go exocom.messageHandler(ws)
	for {
		select {
		case <-exocom.doneCh:
			exocom.doneCh <- true
			log.Println(cyan("EXOCOM: TERMINATING listenToMessages"))
			return
		case incoming := <-exocom.messageCh:
			log.Printf(cyan("EXOCOM: MESSAGE RECEIVED in listenToMessages: %#v\n"), incoming)
		}
	}
}

func (exocom *ExoCom) messageHandler(batman *websocket.Conn) {
	var incoming Message
	for {
		select {
		case <-exocom.doneCh:
			log.Println(cyan("EXOCOM: TERMINATING messageHandler"))
			exocom.doneCh <- true
			return
		default:
			err := websocket.JSON.Receive(batman, &incoming)
			if err == io.EOF {
				log.Println(cyan("EXOCOM: Error:"), red(err))
				log.Println(cyan("EXOCOM: TERMINATING messageHandler"))
				return
			}
			exocom.Lock()
			exocom.saveMessage(incoming)
			exocom.Unlock()
			exocom.messageCh <- incoming
		}
	}
}

func (exocom *ExoCom) saveMessage(message Message) {
	exocom.ReceivedMessages = append(exocom.ReceivedMessages, message)
}

func (exocom *ExoCom) Listen(port int) {
	exocom.port = port

	onConnection := func(ws *websocket.Conn) {
		ws.SetReadDeadline(time.Now().Add(1 * time.Second))
		var incoming Message
		if err := websocket.JSON.Receive(ws, &incoming); err != nil {
			log.Fatal(red(err))
		}
		if incoming.Name == "exocom.register-service" {
			exocom.RegisterService(incoming.Sender, ws)
			exocom.saveMessage(incoming)
			exocom.listenToMessages(ws)
		}
	}

	http.Handle("/services", websocket.Handler(onConnection))
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	log.Println(cyan("===============SERVER TERMINATED==================="))
	if err != nil {
		log.Fatalln(red(err))
	}
	log.Println(cyan("EXOCOM: Listener is done"))
}
