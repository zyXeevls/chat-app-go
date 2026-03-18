package usecase

import "github.com/zyXeevls/chat-app/internal/repository"

type UnreadUseCase struct {
	repo *repository.UnreadRepository
}

func NewUnreadUseCase(repo *repository.UnreadRepository) *UnreadUseCase {
	return &UnreadUseCase{repo: repo}
}

func (u *UnreadUseCase) AddUnread(userID, roomID string) {
	u.repo.Increment(userID, roomID)
}

func (u *UnreadUseCase) ClearUnread(userID, roomID string) {
	u.repo.Reset(userID, roomID)
}

func (u *UnreadUseCase) GetUnread(userID string) ([]map[string]interface{}, error) {
	return u.repo.Get(userID)
}
