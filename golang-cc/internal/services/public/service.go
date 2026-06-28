package public

import (
	"context"
	"strings"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
	"github.com/rafa/golang-cc/internal/utils/email"
)

const SessionDuration = 30 * 24 * time.Hour

type Repository interface {
	FindLoginUser(context.Context, string) (domain.LoginUser, error)
	CreateSession(context.Context, string, string, []byte, time.Time) error
	Authenticate(context.Context, []byte, time.Time) (domain.User, error)
	DeleteSession(context.Context, []byte) error
	InvitationEmail(context.Context, []byte, time.Time) (string, error)
	AcceptInvitation(context.Context, []byte, time.Time, domain.User, string) error
	ActiveUserByEmail(context.Context, string) (string, string, error)
	CreatePasswordReset(context.Context, string, string, []byte, time.Time) error
	ResetPassword(context.Context, []byte, string, time.Time) error
	CreateInvitation(context.Context, string, string, []byte, string, time.Time) error
}

type Service struct {
	repository    Repository
	email         email.Sender
	publicBaseURL string
	now           func() time.Time
}

func New(repository Repository, sender email.Sender, publicBaseURL string) *Service {
	return &Service{repository: repository, email: sender, publicBaseURL: strings.TrimRight(publicBaseURL, "/"), now: time.Now}
}
