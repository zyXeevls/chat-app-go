package websocket

type Hub struct {
	clients map[*Client]bool

	rooms map[string]map[*Client]bool

	register chan *Client

	unregister chan *Client

	broadcast chan *RoomMessage
}

type RoomMessage struct {
	roomID  string
	message []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *RoomMessage),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true

		case client := <-h.unregister:
			delete(h.clients, client)
			for roomID := range h.rooms {
				delete(h.rooms[roomID], client)
			}

		case msg := <-h.broadcast:
			clients := h.rooms[msg.roomID]
			for client := range clients {
				select {
				case client.send <- msg.message:
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
