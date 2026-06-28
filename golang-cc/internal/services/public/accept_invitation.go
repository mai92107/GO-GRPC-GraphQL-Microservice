package public

import (
	"context"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/secure"
)

func (s *Service) AcceptInvitation(ctx context.Context, token, displayName, password string) (domain.User, error) {
	hash := secure.Hash(token)
	email, err := s.repository.InvitationEmail(ctx, hash, s.now())
	if err != nil {
		return domain.User{}, err
	}
	passwordHash, err := HashPassword(password)
	if err != nil {
		return domain.User{}, err
	}
	user := domain.User{ID: secure.UUID(), Email: email, DisplayName: displayName, Role: "member", Status: "active"}
	if err = s.repository.AcceptInvitation(ctx, hash, s.now(), user, passwordHash); err != nil {
		return domain.User{}, err
	}
	return user, nil
}
