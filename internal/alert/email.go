package alert

import (
	"fmt"
	"net/smtp"
	"strings"
)

// EmailNotifier sends alert notifications via SMTP email.
type EmailNotifier struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	To       []string
}

// NewEmailNotifier creates a new EmailNotifier with the given SMTP configuration.
func NewEmailNotifier(host string, port int, username, password, from string, to []string) *EmailNotifier {
	return &EmailNotifier{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		To:       to,
	}
}

// Notify sends an email alert with the given subject and message body.
func (e *EmailNotifier) Notify(subject, message string) error {
	addr := fmt.Sprintf("%s:%d", e.Host, e.Port)

	var auth smtp.Auth
	if e.Username != "" {
		auth = smtp.PlainAuth("", e.Username, e.Password, e.Host)
	}

	body := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		e.From,
		strings.Join(e.To, ", "),
		subject,
		message,
	)

	err := smtp.SendMail(addr, auth, e.From, e.To, []byte(body))
	if err != nil {
		return fmt.Errorf("email notifier: failed to send mail: %w", err)
	}
	return nil
}
