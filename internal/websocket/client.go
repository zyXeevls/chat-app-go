package websocket

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	userID      string
	currentRoom string
	joinedRooms map[string]bool
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

			if !c.joinedRooms[t.RoomID] {
				log.Printf("Unauthorized: User %s tidak bisa typing di room %s", c.userID, t.RoomID)
				continue
			}

			c.hub.broadcast <- &RoomMessage{
				roomID:  t.RoomID,
				message: message,
			}

		case "typing_stop":
			var t TypingEvent

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &t)

			if !c.joinedRooms[t.RoomID] {
				log.Printf("Unauthorized: User %s tidak bisa typing di room %s", c.userID, t.RoomID)
				continue
			}

			c.hub.broadcast <- &RoomMessage{
				roomID:  t.RoomID,
				message: message,
			}

		case "send_message":
			var msg ChatMessage

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &msg)

			if !c.joinedRooms[msg.RoomID] {
				log.Printf("Unauthorized: User %s tidak bisa send message ke room %s", c.userID, msg.RoomID)
				continue
			}

			msg.ID = uuid.New().String()

			newMsg, _ := json.Marshal(map[string]interface{}{
				"event": "receive_message",
				"data": map[string]interface{}{
					"id":       msg.ID,
					"room_id":  msg.RoomID,
					"message":  msg.Message,
					"file_url": msg.FileURL,
					"type":     msg.Type,
					"status":   "sent",
					"sender":   c.userID,
				},
			})

			c.hub.broadcast <- &RoomMessage{
				roomID:   msg.RoomID,
				senderID: c.userID,
				content:  msg.Message,
				fileURL:  msg.FileURL,
				msgType:  msg.Type,
				raw:      newMsg,
			}

		case "message_delivered":
			var data struct {
				MessageID string `json:"message_id"`
			}
			raw, _ := json.Marshal(event.Data)
			json.Unmarshal(raw, &data)

			if c.currentRoom == "" {
				log.Printf("Unauthorized: User %s belum join room", c.userID)
				continue
			}

			go c.hub.messageRepo.UpdateStatus(data.MessageID, "delivered")

			c.hub.broadcast <- &RoomMessage{
				roomID: c.currentRoom,
				raw:    message,
			}

		case "message_read":
			var data struct {
				MessageID string `json:"message_id"`
			}

			raw, _ := json.Marshal(event.Data)
			json.Unmarshal(raw, &data)

			if c.currentRoom == "" {
				log.Printf("Unauthorized: User %s belum join room", c.userID)
				continue
			}

			go c.hub.messageRepo.UpdateStatus(data.MessageID, "read")

			c.hub.broadcast <- &RoomMessage{
				roomID: c.currentRoom,
				raw:    message,
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
