package usecase

import "github.com/zyXeevls/chat-app/internal/repository"

type PresenceUseCase struct {
	repo *repository.PresenceRepository
}

func NewPresenceUseCase(repo *repository.PresenceRepository) *PresenceUseCase {
	return &PresenceUseCase{repo: repo}
}

func (u *PresenceUseCase) SetOnline(userID string) {
	u.repo.SetOnline(userID)
}

func (u *PresenceUseCase) SetOffline(userID string) {
	u.repo.SetOffline(userID)
}

func (u *PresenceUseCase) GetStatus(userID string) (string, int64) {
	if u.repo.IsOnline(userID) {
		return "online", 0
	}
	return "offline", u.repo.GetLastSeen(userID)
}
