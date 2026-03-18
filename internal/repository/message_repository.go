package repository

import (
	"context"

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
	fileURL string,
	msgtype string,
) error {
	query := `
		INSERT INTO messages (id, room_id, sender_id, content, file_url, type, status)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 'sent')
	`
	_, err := r.DB.Exec(context.Background(),
		query,
		roomID,
		senderID,
		content,
		fileURL,
		msgtype,
	)

	return err
}

func (r *MessageRepository) UpdateStatus(messageID string, status string) error {
	query := `
		UPDATE messages
		SET status = $1
		WHERE id = $2
		`

	_, err := r.DB.Exec(context.Background(),
		query,
		status,
		messageID,
	)
	return err
}

// CanUserAccessRoom verifies if user has access to the room
// Returns true if user is a member of the room
func (r *MessageRepository) CanUserAccessRoom(userID, roomID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM room_members 
			WHERE user_id = $1 AND room_id = $2
		) OR EXISTS (
			SELECT 1 FROM rooms 
			WHERE id = $2 AND created_by = $1
		)
	`

	var hasAccess bool
	err := r.DB.QueryRow(context.Background(), query, userID, roomID).Scan(&hasAccess)
	if err != nil {
		return false, err
	}

	return hasAccess, nil
}
