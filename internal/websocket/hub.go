package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/zyXeevls/chat-app/internal/repository"
	"github.com/zyXeevls/chat-app/internal/usecase"
)

type Hub struct {
	clients map[*Client]bool
	rooms   map[string]map[*Client]bool

	register   chan *Client
	unregister chan *Client
	broadcast  chan *RoomMessage

	messageRepo   *repository.MessageRepository
	redis         *redis.Client
	presence      *usecase.PresenceUseCase
	unreadUsecase *usecase.UnreadUseCase
}

type RoomMessage struct {
	RoomID   string `json:"room_id"`
	SenderID string `json:"sender_id,omitempty"`
	Content  string `json:"content,omitempty"`
	FileURL  string `json:"file_url,omitempty"`
	MsgType  string `json:"msg_type,omitempty"`
	Raw      []byte `json:"raw"`
	Persist  bool   `json:"persist"`
}

func NewHub(repo *repository.MessageRepository, redisClient *redis.Client, presenceUseCase *usecase.PresenceUseCase, unreadUseCase *usecase.UnreadUseCase) *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		rooms:         make(map[string]map[*Client]bool),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
		broadcast:     make(chan *RoomMessage),
		messageRepo:   repo,
		redis:         redisClient,
		presence:      presenceUseCase,
		unreadUsecase: unreadUseCase,
	}
}

func (h *Hub) Run() {
	if h.redis != nil {
		go h.subscribe()
	}

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			if h.presence != nil {
				h.presence.SetOnline(client.userID)
			}

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
			if h.presence != nil {
				h.presence.SetOffline(client.userID)
			}

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
			if msg.Persist && msg.SenderID != "" && msg.RoomID != "" && (msg.Content != "" || msg.FileURL != "") {
				err := h.messageRepo.SaveMessage(
					msg.RoomID,
					msg.SenderID,
					msg.Content,
					msg.FileURL,
					msg.MsgType,
				)
				if err != nil {
					log.Printf("Error saving message: %v", err)
				} else {
					log.Printf("Message saved - Room: %s, Sender: %s, Content: %s", msg.RoomID, msg.SenderID, msg.Content)
				}
			}

			if h.redis != nil {
				if err := h.publish(*msg); err != nil {
					log.Printf("Redis publish failed, fallback local broadcast: %v", err)
					h.dispatch(msg)
				}
				continue
			}

			h.dispatch(msg)
		}
	}
}

func (h *Hub) dispatch(msg *RoomMessage) {
	clients := h.rooms[msg.RoomID]

	for client := range clients {
		select {
		case client.send <- msg.Raw:
		default:
			close(client.send)
			delete(clients, client)
		}
	}
}

func (h *Hub) publish(msg RoomMessage) error {
	data, _ := json.Marshal(msg)
	return h.redis.Publish(context.Background(), "chat", data).Err()
}

func (h *Hub) subscribe() {
	pubsub := h.redis.Subscribe(context.Background(), "chat")

	for {
		msg, err := pubsub.ReceiveMessage(context.Background())
		if err != nil {
			log.Printf("Redis subscribe receive error: %v", err)
			continue
		}

		var rm RoomMessage
		if err := json.Unmarshal([]byte(msg.Payload), &rm); err != nil {
			log.Printf("Redis payload parse error: %v", err)
			continue
		}

		h.dispatch(&rm)
	}
}

func (h *Hub) JoinRoom(roomRef string, client *Client) (string, error) {
	roomID, err := h.messageRepo.ResolveRoomID(roomRef)
	if err != nil {
		log.Printf("Error resolving room %s for user %s: %v", roomRef, client.userID, err)
		return "", fmt.Errorf("failed to resolve room: %w", err)
	}

	if roomID == "" {
		roomID = roomRef
	}

	canAccess, err := h.messageRepo.CanUserAccessRoom(client.userID, roomID)
	if err != nil {
		log.Printf("Error checking room access for user %s in room %s: %v", client.userID, roomID, err)
		return "", fmt.Errorf("failed to verify room access: %w", err)
	}

	if !canAccess {
		provisionedRoomID, err := h.messageRepo.EnsureUserRoomAccess(client.userID, roomRef)
		if err != nil {
			log.Printf("Access denied: User %s tidak bisa join room %s", client.userID, roomID)
			return "", fmt.Errorf("access denied for room %s", roomRef)
		}

		roomID = provisionedRoomID

		canAccess, err = h.messageRepo.CanUserAccessRoom(client.userID, roomID)
		if err != nil {
			return "", fmt.Errorf("failed to verify room access after provision: %w", err)
		}

		if !canAccess {
			log.Printf("Access denied after provision: User %s tidak bisa join room %s", client.userID, roomID)
			return "", fmt.Errorf("access denied for room %s", roomID)
		}
	}

	if h.rooms[roomID] == nil {
		h.rooms[roomID] = make(map[*Client]bool)
	}

	h.rooms[roomID][client] = true
	client.joinedRooms[roomID] = true
	client.currentRoom = roomID

	return roomID, nil
}
