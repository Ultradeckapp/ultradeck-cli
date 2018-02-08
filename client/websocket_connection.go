package client

import (
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 1024 * 1024
)

const RegisterListenerRequest = "register_listener"
const OkResponse = "ok"

type Request struct {
	ClientID string                 `json:"client_id"`
	Channel  string                 `json:"channel"`
	Request  string                 `json:"request"`
	Data     map[string]interface{} `json:"data"`
}

// Client is the client
type WebsocketConnection struct {
	Conn      *websocket.Conn
	Done      chan bool
	Interrupt chan os.Signal
	ClientID  string
}

func NewWebsocketConnection() *WebsocketConnection {
	ws := &WebsocketConnection{ClientID: NewUUID()}
	ws.OpenConnection()
	return ws
}

func (c *WebsocketConnection) OpenConnection() {
	c.Interrupt = make(chan os.Signal, 1)
	signal.Notify(c.Interrupt, os.Interrupt)

	log.Printf("connecting to %s", c.serverURL())

	var err error
	DebugMsg("Dialing...")
	c.Conn, _, err = websocket.DefaultDialer.Dial(c.serverURL(), nil)
	if err != nil {
		log.Fatal("Dial err:", err)
	}
	DebugMsg("Dialed")

	c.Done = make(chan bool)
}

func (c *WebsocketConnection) CloseConnection() {
	err := c.Conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("write close err:", err)
		return
	}
	c.Conn.Close()
	close(c.Done)
}

func (c *WebsocketConnection) RegisterListener(channel string) {
	req := &Request{
		Request:  RegisterListenerRequest,
		ClientID: c.ClientID,
		Channel:  channel,
	}
	authMsg, _ := json.Marshal(req)

	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(authMsg))
	if err != nil {
		log.Println("write err: ", err)
	}
}

func (c *WebsocketConnection) Listen(rchan chan<- *Request) {
	log.Println("Listening..")
	for {
		_, message, err := c.Conn.ReadMessage()
		DebugMsg("<Websocket> read message")

		if err != nil {
			c.Done <- true
			log.Println("read error:", err)
			break
		}

		req := &Request{}
		json.Unmarshal(message, req)
		rchan <- req
	}
}

func (c *WebsocketConnection) serverURL() string {
	addr := "localhost:8080"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	return u.String()
}
