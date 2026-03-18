package usecase

import "github.com/zyXeevls/chat-app/internal/repository"

type MessageUseCase struct {
	repo repository.MessageRepository
}

func (u *MessageUseCase) GetMessages(roomID string, page int, limit int) ([]repository.Message, error) {
	return u.repo.GetMessage(roomID, "", page, limit)
}
