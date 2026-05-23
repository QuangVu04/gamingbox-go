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
		"MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n" +
		body)

	auth := smtp.PlainAuth("", from, password, smtpHost)

	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
	if err != nil {
		return err
	}
	return nil
}

// GenerateOTPEmailTemplate generates a beautifully designed HTML template for OTP emails.
func GenerateOTPEmailTemplate(title, message, otp string) string {
	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
</head>
<body style="margin: 0; padding: 0; font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif; background-color: #0f1419; color: #ffffff;">
    <div style="max-w-md; margin: 0 auto; background-color: #1b2838; padding: 40px; border-radius: 16px; margin-top: 40px; box-shadow: 0 10px 25px rgba(0,0,0,0.5); text-align: center; border: 1px solid rgba(255,255,255,0.05);">
        <h1 style="color: #67c1f5; font-size: 28px; margin-bottom: 10px; font-weight: 800; letter-spacing: -0.5px;">GamingBox</h1>
        <h2 style="font-size: 20px; font-weight: 600; margin-bottom: 20px;">%s</h2>
        
        <p style="color: #94a3b8; font-size: 15px; margin-bottom: 30px; line-height: 1.6;">
            %s
        </p>

        <div style="background-color: #0f1419; padding: 25px; border-radius: 12px; margin-bottom: 30px; border: 1px solid rgba(255,255,255,0.1);">
            <div style="font-size: 36px; font-weight: 900; letter-spacing: 8px; color: #ffffff;">%s</div>
        </div>

        <p style="color: #64748b; font-size: 13px; margin-bottom: 0;">
            Mã này sẽ hết hạn trong vòng 5 phút.<br/>
            Nếu bạn không yêu cầu mã này, vui lòng bỏ qua email.
        </p>
    </div>
</body>
</html>
`, title, title, message, otp)
}
