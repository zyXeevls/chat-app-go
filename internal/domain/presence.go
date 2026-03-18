package domain

type Presence struct {
	UserID   string `json:"user_id"`
	Status   string `json:"status"`
	LastSeen int64  `json:"last_seen"`
}
