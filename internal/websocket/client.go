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

			if room.RoomID == "" {
				resp, _ := json.Marshal(map[string]interface{}{
					"event": "join_room_error",
					"data": map[string]interface{}{
						"room_id": room.RoomID,
						"reason":  "room_id is required",
					},
				})
				c.send <- resp
				continue
			}

			joinedRoomID, err := c.hub.JoinRoom(room.RoomID, c)
			if err != nil {
				resp, _ := json.Marshal(map[string]interface{}{
					"event": "join_room_error",
					"data": map[string]interface{}{
						"room_id": room.RoomID,
						"reason":  err.Error(),
					},
				})
				c.send <- resp
				continue
			}

			okResp, _ := json.Marshal(map[string]interface{}{
				"event": "join_room_ok",
				"data": map[string]interface{}{
					"room_id":          room.RoomID,
					"resolved_room_id": joinedRoomID,
				},
			})
			c.send <- okResp

		case "typing_start":
			var t TypingEvent

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &t)

			if !c.joinedRooms[t.RoomID] {
				log.Printf("Unauthorized: User %s tidak bisa typing di room %s", c.userID, t.RoomID)
				continue
			}

			c.hub.broadcast <- &RoomMessage{
				RoomID:  t.RoomID,
				Raw:     message,
				Persist: false,
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
				RoomID:  t.RoomID,
				Raw:     message,
				Persist: false,
			}

		case "send_message":
			var msg ChatMessage

			data, _ := json.Marshal(event.Data)
			json.Unmarshal(data, &msg)

			if !c.joinedRooms[msg.RoomID] {
				joinedRoomID, err := c.hub.JoinRoom(msg.RoomID, c)
				if err != nil {
					log.Printf("Unauthorized: User %s tidak bisa send message ke room %s: %v", c.userID, msg.RoomID, err)

					resp, _ := json.Marshal(map[string]interface{}{
						"event": "send_message_error",
						"data": map[string]interface{}{
							"room_id": msg.RoomID,
							"reason":  err.Error(),
						},
					})
					c.send <- resp
					continue
				}

				msg.RoomID = joinedRoomID
			}

			if msg.Message == "" {
				msg.Message = msg.Content
			}

			if msg.Type == "" {
				if msg.FileURL != "" {
					msg.Type = "file"
				} else {
					msg.Type = "text"
				}
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
				RoomID:   msg.RoomID,
				SenderID: c.userID,
				Content:  msg.Message,
				FileURL:  msg.FileURL,
				MsgType:  msg.Type,
				Raw:      newMsg,
				Persist:  true,
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
				RoomID:  c.currentRoom,
				Raw:     message,
				Persist: false,
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
				RoomID:  c.currentRoom,
				Raw:     message,
				Persist: false,
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
