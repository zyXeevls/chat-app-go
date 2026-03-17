package websocket

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub *Hub

	com *websocket.Conn

	send chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.com.Close()
	}()
	for {
		_, message, err := c.com.ReadMessage()

		if err != nil {
			log.Println(err)
			break
		}

		var event Event

		err = json.Unmarshal(message, &event)

		if err != nil {
			log.Println("invalid event")
			continue
		}

		switch event.Event {
		case "send_message":
			c.hub.broadcast <- message
		default:
			log.Println("unknown event", event.Event)
		}
	}
}

func (c *Client) writePump() {
	defer c.com.Close()
	for {
		message, ok := <-c.send

		if !ok {
			c.com.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.com.WriteMessage(websocket.TextMessage, message)

		if err != nil {
			return
		}
	}
}
