package public

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rafa/golang-cc/internal/domain"
)

type fakeRepository struct {
	loginUser      domain.LoginUser
	loginErr       error
	sessionCreated bool
	activeUserErr  error
	resetCreated   bool
}

func (f *fakeRepository) FindLoginUser(context.Context, string) (domain.LoginUser, error) {
	return f.loginUser, f.loginErr
}
func (f *fakeRepository) CreateSession(context.Context, string, string, []byte, time.Time) error {
	f.sessionCreated = true
	return nil
}
func (f *fakeRepository) Authenticate(context.Context, []byte, time.Time) (domain.User, error) {
	return domain.User{}, nil
}
func (f *fakeRepository) DeleteSession(context.Context, []byte) error { return nil }
func (f *fakeRepository) InvitationEmail(context.Context, []byte, time.Time) (string, error) {
	return "", nil
}
func (f *fakeRepository) AcceptInvitation(context.Context, []byte, time.Time, domain.User, string) error {
	return nil
}
func (f *fakeRepository) ActiveUserByEmail(context.Context, string) (string, string, error) {
	return "", "", f.activeUserErr
}
func (f *fakeRepository) CreatePasswordReset(context.Context, string, string, []byte, time.Time) error {
	f.resetCreated = true
	return nil
}
func (f *fakeRepository) ResetPassword(context.Context, []byte, string, time.Time) error { return nil }
func (f *fakeRepository) CreateInvitation(context.Context, string, string, []byte, string, time.Time) error {
	return nil
}

type fakeSender struct{}

func (fakeSender) SendInvitation(context.Context, string, string) error    { return nil }
func (fakeSender) SendPasswordReset(context.Context, string, string) error { return nil }

func TestLoginCreatesSessionForValidActiveUser(t *testing.T) {
	password := "valid-password-123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	repository := &fakeRepository{loginUser: domain.LoginUser{
		User: domain.User{ID: "user-id", Status: "active"}, PasswordHash: hash,
	}}
	service := New(repository, fakeSender{}, "http://localhost")

	if _, token, err := service.Login(context.Background(), "member@example.test", password); err != nil || token == "" {
		t.Fatalf("login token=%q err=%v", token, err)
	}
	if !repository.sessionCreated {
		t.Fatal("login did not create session")
	}
}

func TestRequestPasswordResetDoesNotRevealMissingUser(t *testing.T) {
	repository := &fakeRepository{activeUserErr: domain.ErrNotFound}
	service := New(repository, fakeSender{}, "http://localhost")

	if err := service.RequestPasswordReset(context.Background(), "missing@example.test"); err != nil {
		t.Fatal(err)
	}
	if repository.resetCreated {
		t.Fatal("password reset token created for missing user")
	}
}

func TestLoginRejectsRepositoryError(t *testing.T) {
	service := New(&fakeRepository{loginErr: errors.New("database unavailable")}, fakeSender{}, "http://localhost")
	if _, _, err := service.Login(context.Background(), "member@example.test", "valid-password-123"); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
}
