package gmail

import (
	"context"
	"fmt"
	"mime"
	"net/smtp"
	"strings"

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

func (g *Gmail) Send(ctx context.Context, to, subject, plain, html string) error {
	logger := loggerPkg.From(ctx)

	auth := smtp.PlainAuth("", g.username, g.password, g.host)

	var msg strings.Builder
	encodedSubject := mime.QEncoding.Encode("utf-8", subject)
	msg.WriteString(fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n", g.username, to, encodedSubject))

	if html != "" {
		msg.WriteString("MIME-Version: 1.0\r\n")
		msg.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		msg.WriteString(html)
	} else {
		msg.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
		msg.WriteString(plain)
	}

	smtpAddr := fmt.Sprintf("%s:%s", g.host, g.port)
	if err := smtp.SendMail(smtpAddr, auth, g.username, []string{to}, []byte(msg.String())); err != nil {
		logger.Error("Gmail SMTP failed", "err", err)
		return fmt.Errorf("gmail send failed: %w", err)
	}

	logger.Debug("Email sent via Gmail",
		"to", to,
		"subject", subject,
		"has_html", html != "")
	return nil
}
