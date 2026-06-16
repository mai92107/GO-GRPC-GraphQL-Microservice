package email

import (
	"context"
	"testing"

	"github.com/rafa/golang-cc/internal/platform/config"
)

func TestSMTPDeliversInvitationAndReset(t *testing.T) {
	cfg, err := config.LoadFromProject("configs/test.json")
	if err != nil {
		t.Fatal(err)
	}
	sender := SMTP{Addr: cfg.Mail.SMTPAddress, From: cfg.Mail.From}
	if err := sender.SendInvitation(context.Background(), "invite@example.test", "http://localhost/accept-invitation?token=test"); err != nil {
		t.Skipf("Mail server unavailable: %v", err)
	}
	if err := sender.SendPasswordReset(context.Background(), "reset@example.test", "http://localhost/reset-password?token=test"); err != nil {
		t.Fatal(err)
	}
}
