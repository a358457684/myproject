// Package ws is to define a websocket server and client connect.

package websocket

import (
	"common/log"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
)

const (
	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10
)

// ClientManager is a websocket manager
type ClientManager struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
}

// Client is a websocket client
type Client struct {
	ID     string
	Socket *websocket.Conn
	Send   chan []byte
	Data   interface{}
}

// Message is an object for websocket message which is mapped to json type
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

// Manager define a ws server manager
var Manager = ClientManager{
	Broadcast:  make(chan []byte),
	Register:   make(chan *Client),
	Unregister: make(chan *Client),
	Clients:    make(map[*Client]bool),
}

var wslock = sync.RWMutex{}

func Init() {
	Manager.Start()
}

// Start is to start a ws server
func (manager *ClientManager) Start() {
	for {
		select {
		case conn := <-manager.Register:
			wslock.Lock()
			manager.Clients[conn] = true
			wslock.Unlock()

		case conn := <-manager.Unregister:
			wslock.Lock()
			if _, ok := manager.Clients[conn]; ok {
				close(conn.Send)
				delete(manager.Clients, conn)
			}
			wslock.Unlock()
		case message := <-manager.Broadcast:
			wslock.Lock()
			for conn := range manager.Clients {
				select {
				case conn.Send <- message:
				default:
					close(conn.Send)
					delete(manager.Clients, conn)
				}
			}
			wslock.Unlock()
		}

	}
}

func (c *Client) Read() {
	defer func() {

		Manager.Unregister <- c
		c.Socket.Close()
	}()

	for {
		messageType, data, err := c.Socket.ReadMessage()
		if err != nil {

			Manager.Unregister <- c
			c.Socket.Close()
			break
		}
		// 更换数据
		if messageType == websocket.TextMessage && data != nil {
			_ = json.Unmarshal(data, &c.Data)
		}
	}
}

func (c *Client) Write() {
	defer func() {
		c.Socket.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Socket.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			c.Socket.WriteMessage(websocket.TextMessage, message)
		}
	}
}

func Notify(info interface{}) {
	defer func() {
		err := recover() // recover() 捕获panic异常，获得程序执行权。
		if err != nil {
			log.Error("ws Notify", err, info)
		}
	}()

	jsonMessage, _ := json.Marshal(&info)
	Manager.Broadcast <- jsonMessage
}

func GetUpgrade(protocol string) websocket.Upgrader {
	return websocket.Upgrader{
		// cross origin domain
		CheckOrigin: func(r *http.Request) bool { //这个是解决跨域问题
			return true
		},
		Subprotocols: []string{protocol},
		//将获取的参数放进这个数组，问题解决
	}
}
