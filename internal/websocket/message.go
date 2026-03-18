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
	ID      string `json:"id"`
	RoomID  string `json:"room_id"`
	Message string `json:"message"`
	Content string `json:"content,omitempty"`
	FileURL string `json:"file_url,omitempty"`
	Type    string `json:"type"`
}

type TypingEvent struct {
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
}
