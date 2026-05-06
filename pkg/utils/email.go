package utils

import (
	"fmt"
	"net/smtp"
	"vault/be/config"
)

// SendEmail sends a plain text email via SMTP
func SendEmail(to string, subject string, body string) error {
	from := config.App.SMTPUser
	password := config.App.SMTPPassword
	smtpHost := config.App.SMTPHost
	smtpPort := config.App.SMTPPort

	if from == "" || password == "" {
		return fmt.Errorf("SMTP credentials not configured")
	}

	message := []byte("Subject: " + subject + "\r\n" +
		"To: " + to + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n" +
		body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return err
	}
	return nil
}
