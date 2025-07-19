package gmail

import (
	"fmt"
	"net/smtp"
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

func (g *Gmail) Send(to, subject, _ string, html string) error {
	addr := g.host + ":" + g.port

	auth := smtp.PlainAuth("", g.username, g.password, g.host)

	msg := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n%s",
		g.username,
		to,
		subject,
		html,
	)

	return smtp.SendMail(addr, auth, g.username, []string{to}, []byte(msg))
}
