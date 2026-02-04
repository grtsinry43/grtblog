package email

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/grtsinry43/grtblog-v2/server/internal/app/sysconfig"
	domainemail "github.com/grtsinry43/grtblog-v2/server/internal/domain/email"
)

type Message struct {
	To       []string
	Subject  string
	HTMLBody string
	TextBody string
}

type Sender struct {
	sysCfg *sysconfig.Service
}

func NewSender(sysCfg *sysconfig.Service) *Sender {
	return &Sender{sysCfg: sysCfg}
}

func (s *Sender) Send(ctx context.Context, msg Message) error {
	settings, err := s.sysCfg.EmailSettings(ctx)
	if err != nil {
		return err
	}
	if !settings.Enabled {
		return domainemail.ErrEmailDisabled
	}
	to := normalizeRecipients(msg.To)
	if len(to) == 0 {
		to = normalizeRecipients(settings.DefaultTo)
	}
	if len(to) == 0 {
		return domainemail.ErrEmailNoRecipient
	}
	if strings.TrimSpace(settings.FromAddress) == "" || strings.TrimSpace(settings.SMTPHost) == "" || settings.SMTPPort <= 0 {
		return domainemail.ErrEmailConfigInvalid
	}
	if strings.TrimSpace(msg.Subject) == "" {
		return domainemail.ErrEmailTemplateRenderFailed
	}
	raw := buildMessage(settings.FromName, settings.FromAddress, to, msg)
	if err := sendSMTP(ctx, settings, settings.FromAddress, to, raw); err != nil {
		return fmt.Errorf("%w: %v", domainemail.ErrEmailSendFailed, err)
	}
	return nil
}

func sendSMTP(ctx context.Context, settings sysconfig.EmailSettings, from string, to []string, payload []byte) error {
	address := fmt.Sprintf("%s:%d", settings.SMTPHost, settings.SMTPPort)
	dialer := net.Dialer{Timeout: settings.Timeout}

	var conn net.Conn
	var err error
	if settings.TLSMode == sysconfig.EmailTLSModeTLS {
		conn, err = tls.DialWithDialer(&dialer, "tcp", address, &tls.Config{ServerName: settings.SMTPHost})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", address)
	}
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, settings.SMTPHost)
	if err != nil {
		return err
	}
	defer client.Close()

	if settings.TLSMode == sysconfig.EmailTLSModeStartTLS {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return fmt.Errorf("smtp server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: settings.SMTPHost}); err != nil {
			return err
		}
	}

	if strings.TrimSpace(settings.SMTPUsername) != "" {
		auth := smtp.PlainAuth("", settings.SMTPUsername, settings.SMTPPassword, settings.SMTPHost)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(payload); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	if err := client.Quit(); err != nil {
		return err
	}
	return nil
}

func buildMessage(fromName, fromAddress string, to []string, msg Message) []byte {
	from := strings.TrimSpace(fromAddress)
	if strings.TrimSpace(fromName) != "" {
		from = fmt.Sprintf("%s <%s>", strings.TrimSpace(fromName), strings.TrimSpace(fromAddress))
	}
	boundary := fmt.Sprintf("grtblog-%d", time.Now().UnixNano())
	buf := bytes.NewBuffer(nil)
	buf.WriteString("From: " + from + "\r\n")
	buf.WriteString("To: " + strings.Join(to, ",") + "\r\n")
	buf.WriteString("Subject: " + encodeSubject(msg.Subject) + "\r\n")
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n")
	buf.WriteString("\r\n")

	textBody := msg.TextBody
	if strings.TrimSpace(textBody) == "" {
		textBody = stripHTML(msg.HTMLBody)
	}
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n\r\n")
	buf.WriteString(textBody)
	buf.WriteString("\r\n")

	htmlBody := msg.HTMLBody
	if strings.TrimSpace(htmlBody) == "" {
		htmlBody = "<pre>" + textBody + "</pre>"
	}
	buf.WriteString("--" + boundary + "\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
	buf.WriteString(htmlBody)
	buf.WriteString("\r\n")
	buf.WriteString("--" + boundary + "--\r\n")
	return buf.Bytes()
}

func encodeSubject(subject string) string {
	return strings.ReplaceAll(subject, "\n", " ")
}

func stripHTML(html string) string {
	replacer := strings.NewReplacer("<br>", "\n", "<br/>", "\n", "<br />", "\n", "</p>", "\n", "<p>", "")
	plain := replacer.Replace(html)
	plain = strings.ReplaceAll(plain, "<", "")
	plain = strings.ReplaceAll(plain, ">", "")
	return strings.TrimSpace(plain)
}

func normalizeRecipients(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		for _, part := range strings.Split(item, ",") {
			email := strings.TrimSpace(part)
			if email == "" {
				continue
			}
			if _, ok := seen[email]; ok {
				continue
			}
			seen[email] = struct{}{}
			result = append(result, email)
		}
	}
	return result
}
