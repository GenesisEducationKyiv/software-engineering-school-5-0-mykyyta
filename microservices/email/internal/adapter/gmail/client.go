package gmail

import (
	"context"
	"fmt"
	"net/smtp"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type Gmail struct {
	username string
	password string
	host     string
	port     string
}

func New(username, password string) *Gmail {
	return &Gmail{
		username: username,
		password: password,
		host:     "smtp.gmail.com",
		port:     "587",
	}
}

func (g *Gmail) Send(ctx context.Context, to, subject, _ string, html string) error {
	loggerPkg.From(ctx).Info("Sending email via Gmail SMTP", "to", to, "host", g.host)

	addr := g.host + ":" + g.port

	auth := smtp.PlainAuth("", g.username, g.password, g.host)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		g.username,
		to,
		subject,
		html,
	)

	if err := smtp.SendMail(addr, auth, g.username, []string{to}, []byte(msg)); err != nil {
		loggerPkg.From(ctx).Error("Gmail SMTP send failed", "to", to, "host", addr, "err", err)
		return err
	}

	loggerPkg.From(ctx).Info("Email sent successfully via Gmail", "to", to)
	return nil
}
