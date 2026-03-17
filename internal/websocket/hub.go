package websocket

import (
	"log"

	"github.com/zyXeevls/chat-app/internal/repository"
)

type Hub struct {
	clients map[*Client]bool
	rooms   map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan *RoomMessage

	messageRepo *repository.MessageRepository
}

type RoomMessage struct {
	roomID   string
	senderID string
	content  string
	raw      []byte
	message  []byte
}

func NewHub(repo *repository.MessageRepository) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		rooms:       make(map[string]map[*Client]bool),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		broadcast:   make(chan *RoomMessage),
		messageRepo: repo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

			msg := []byte(`{
				"event":"user_online",
				"data":{"user_id":"` + client.userID + `"}
			}`)

			for c := range h.clients {
				if c != client {
					c.send <- msg
				}
			}

		case client := <-h.unregister:
			delete(h.clients, client)

			msg := []byte(`{
				"event":"user_offline",
				"data":{"user_id":"` + client.userID + `"}
			}`)

			for c := range h.clients {
				if c != client {
					c.send <- msg
				}
			}

			for roomID := range h.rooms {
				delete(h.rooms[roomID], client)
			}

		case msg := <-h.broadcast:
			clients := h.rooms[msg.roomID]

			if msg.senderID != "" && msg.content != "" {
				err := h.messageRepo.SaveMessage(
					msg.roomID,
					msg.senderID,
					msg.content,
					"text",
				)
				if err != nil {
					log.Printf("Error saving message: %v", err)
				} else {
					log.Printf("Message saved - Room: %s, Sender: %s, Content: %s", msg.roomID, msg.senderID, msg.content)
				}
			}

			for client := range clients {
				select {
				case client.send <- msg.raw:
				default:
					close(client.send)
					delete(clients, client)
				}
			}
		}
	}
}

func (h *Hub) JoinRoom(roomID string, client *Client) {
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}

	h.rooms[roomID][client] = true
}
