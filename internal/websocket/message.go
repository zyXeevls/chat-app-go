package websocket

type Event struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type Message struct {
	RoomID  string `json:"room_id"`
	Message string `json:"message"`
}

type ChatMessage struct {
	RoomID  string `json:"room_id"`
	Message string `json:"message"`
}

type TypingEvent struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}
