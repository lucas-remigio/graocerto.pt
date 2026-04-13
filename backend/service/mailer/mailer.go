package mailer

import (
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"
)

type Message struct {
	To      string
	Subject string
	Body    string
}

type Mailer interface {
	Send(message Message) error
}

type NoopMailer struct{}

func (NoopMailer) Send(Message) error {
	return nil
}

type SMTPMailer struct {
	host      string
	port      string
	username  string
	password  string
	fromEmail string
	fromName  string
}

func NewSMTPMailer(host, port, username, password, fromEmail, fromName string) *SMTPMailer {
	return &SMTPMailer{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

func (m *SMTPMailer) Send(message Message) error {
	if m == nil {
		return fmt.Errorf("mailer is not configured")
	}

	addr := fmt.Sprintf("%s:%s", m.host, m.port)
	auth := smtp.PlainAuth("", m.username, m.password, m.host)
	raw := buildMessage(m.fromAddress(), message)

	err := smtp.SendMail(addr, auth, m.fromEmail, []string{message.To}, []byte(raw))
	if err != nil {
		slog.Error("failed to send email", "to", message.To, "subject", message.Subject, "error", err)
		return err
	}

	slog.Info("email sent successfully", "to", message.To, "subject", message.Subject)
	return nil
}

func (m *SMTPMailer) fromAddress() string {
	if m.fromName == "" {
		return m.fromEmail
	}

	return fmt.Sprintf("%s <%s>", m.fromName, m.fromEmail)
}

func buildMessage(from string, message Message) string {
	var builder strings.Builder
	builder.WriteString("From: ")
	builder.WriteString(from)
	builder.WriteString("\r\n")
	builder.WriteString("To: ")
	builder.WriteString(message.To)
	builder.WriteString("\r\n")
	builder.WriteString("Subject: ")
	builder.WriteString(message.Subject)
	builder.WriteString("\r\n")
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	builder.WriteString("\r\n")
	builder.WriteString(message.Body)
	builder.WriteString("\r\n")
	return builder.String()
}
