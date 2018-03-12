package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/skratchdot/open-golang/open"
	"github.com/twinj/uuid"

	"github.com/gammons/ultradeck-cli/ultradeck"
	"github.com/gorilla/websocket"
)

// Client is the client
type WebsocketConnection struct {
	Conn      *websocket.Conn
	Done      chan struct{}
	Interrupt chan os.Signal
}

func (c *WebsocketConnection) OpenConnection() {
	c.Interrupt = make(chan os.Signal, 1)
	signal.Notify(c.Interrupt, os.Interrupt)

	log.Printf("connecting to %s", c.serverURL())

	var err error
	log.Println("Dialing...")
	c.Conn, _, err = websocket.DefaultDialer.Dial(c.serverURL(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("Dialed")

	c.Done = make(chan struct{})

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

// DoAuth does auth
func (c *WebsocketConnection) DoAuth(processMsg func(req *ultradeck.Request)) {
	c.OpenConnection()

	var auth = make(map[string]interface{})
	auth["token"] = uuid.NewV4()
	auth["tokenType"] = "intermediate"

	req := &ultradeck.Request{Request: ultradeck.StartAuthRequest, Data: auth}
	authMsg, _ := json.Marshal(req)

	err := c.Conn.WriteMessage(websocket.TextMessage, []byte(authMsg))
	if err != nil {
		log.Println("write err: ", err)
	}

	url := fmt.Sprintf("http://localhost:3000/start_auth?token=%s", auth["token"])
	open.Start(url)

	c.listen(processMsg)
}

func (c *WebsocketConnection) listen(processMsg func(req *ultradeck.Request)) {
	go func() {
		log.Println("Listening..")
		for {
			_, message, err := c.Conn.ReadMessage()
			if err != nil {
				log.Println("read error:", err)
				break
			}

			req := &ultradeck.Request{}
			json.Unmarshal(message, req)
			processMsg(req)
		}

		c.Conn.Close()
		_, ok := <-c.Done
		if ok {
			close(c.Done)
		}
	}()

	log.Println("after setupMessageReader")

	select {
	case <-c.Done:
		log.Println("Got done msg")
	case <-c.Interrupt:
		c.CloseConnection()
		log.Println("interrupt")
	}
}

func (c *WebsocketConnection) serverURL() string {
	addr := "localhost:8080"
	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	return u.String()
}
