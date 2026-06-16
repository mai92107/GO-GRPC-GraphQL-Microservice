package email

import (
	"context"
	"fmt"
	"net/smtp"
)

type Sender interface {
	SendInvitation(context.Context, string, string) error
	SendPasswordReset(context.Context, string, string) error
}

type SMTP struct {
	Addr string
	From string
}

func (s SMTP) SendInvitation(ctx context.Context, to, link string) error {
	return s.send(ctx, to, "啟用您的信用卡推薦帳號",
		"請使用以下連結啟用帳號並設定密碼：\r\n\r\n"+link)
}

func (s SMTP) SendPasswordReset(ctx context.Context, to, link string) error {
	return s.send(ctx, to, "重設您的信用卡推薦帳號密碼",
		"請使用以下連結重設密碼：\r\n\r\n"+link)
}

func (s SMTP) send(ctx context.Context, to, subject, body string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	message := []byte("From: " + s.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + body + "\r\n")
	if err := smtp.SendMail(s.Addr, nil, s.From, []string{to}, message); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
