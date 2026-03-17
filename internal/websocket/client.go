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
		case "join_room":
			var room struct {
				RoomID string `json:"room_id"`
			}

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &room)
			c.hub.JoinRoom(room.RoomID, c)

		case "send_message":
			var roomMessage RoomMessage
			err = json.Unmarshal(message, &roomMessage)
			if err != nil {
				log.Println("invalid message format")
				continue
			}
			c.hub.broadcast <- &roomMessage
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
