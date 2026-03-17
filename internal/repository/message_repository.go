package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MessageRepository struct {
	DB *pgxpool.Pool
}

func NewMessageRepository(db *pgxpool.Pool) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (r *MessageRepository) SaveMessage(
	roomID string,
	senderID string,
	content string,
	msgtype string,
) error {
	query := `
		INSERT INTO messages (id, room_id, sender_id, content, type)
		VALUES (gen_random_uuid(), @room_id, @sender_id, @content, @msg_type)
	`
	_, err := r.DB.Exec(context.Background(), query, pgx.NamedArgs{
		"room_id":   roomID,
		"sender_id": senderID,
		"content":   content,
		"msg_type":  msgtype,
	})

	return err
}
