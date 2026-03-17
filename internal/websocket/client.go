package websocket

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	userID string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()

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

		case "typing_start":
			var t TypingEvent

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &t)

			c.hub.broadcast <- &RoomMessage{
				roomID:  t.RoomID,
				message: message,
			}

		case "typing_stop":
			var t TypingEvent

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &t)

			c.hub.broadcast <- &RoomMessage{
				roomID:  t.RoomID,
				message: message,
			}

		case "send_message":
			var msg ChatMessage

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &msg)

			c.hub.broadcast <- &RoomMessage{
				roomID:   msg.RoomID,
				senderID: c.userID,
				content:  msg.Message,
				raw:      message,
			}
		default:
			log.Println("unknown event", event.Event)
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for {
		message, ok := <-c.send

		if !ok {
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		err := c.conn.WriteMessage(websocket.TextMessage, message)

		if err != nil {
			return
		}
	}
}
