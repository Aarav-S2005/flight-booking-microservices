package email

import (
	"net/smtp"
)

type EmailSender interface {
	Send(to, subject, body string) error
}

type Sender struct {
	Username string
	Password string
}

func NewSender(username, password string) *Sender {
	return &Sender{username, password}
}

func (s *Sender) Send(to, subject, body string) error {
	auth := smtp.PlainAuth(
		"",
		s.Username,
		s.Password,
		"smtp.gmail.com",
	)

	msg := []byte(
    "From: " + s.Username + "\r\n" +
        "To: " + to + "\r\n" +
        "Subject: " + subject + "\r\n" +
        "MIME-Version: 1.0\r\n" +
        "Content-Type: text/html; charset=UTF-8\r\n" +
        "\r\n" +
        body,
	)

	return smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		s.Username,
		[]string{to},
		msg,
	)
}
