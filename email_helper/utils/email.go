package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	gomail "gopkg.in/mail.v2"
)

type EmailAttachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

type EmailOptions struct {
	From        string
	To          string
	Subject     string
	HTML        string
	Attachments []EmailAttachment
}

var (
	smtpMu     sync.Mutex
	smtpDialer *gomail.Dialer
	smtpConn   gomail.SendCloser
)

const (
	defaultSMTPHost = "smtp.gmail.com"
	defaultSMTPPort = 587
	defaultSMTPUser = "contact@newyolk.io"
	defaultSMTPPass = "xtwhlshqszeaonfg"
)

// deriveFromAddress returns the fixed sender.
func deriveFromAddress() string {
	return defaultSMTPUser
}

func smtpTimeout() time.Duration {
	if v := strings.TrimSpace(os.Getenv("EMAIL_TIMEOUT_SECONDS")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 20 * time.Second
}

func tlsConfigForHost(host string) *tls.Config {
	return &tls.Config{
		ServerName:         host,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: strings.EqualFold(strings.TrimSpace(os.Getenv("EMAIL_TLS_INSECURE")), "true"),
	}
}

func getSMTPDialer() (*gomail.Dialer, error) {
	// Fixed SMTP settings as requested (no env lookup)
	host := defaultSMTPHost
	port := defaultSMTPPort
	user := defaultSMTPUser
	pass := defaultSMTPPass

	d := gomail.NewDialer(host, port, user, pass)
	d.Timeout = smtpTimeout()
	d.LocalName = "localhost"
	d.SSL = port == 465
	d.TLSConfig = tlsConfigForHost(host)
	return d, nil
}

func ensureSMTPConnLocked() (gomail.SendCloser, error) {
	if smtpConn != nil {
		return smtpConn, nil
	}

	d, err := getSMTPDialer()
	if err != nil {
		return nil, err
	}
	smtpDialer = d

	c, err := d.Dial()
	if err != nil {
		return nil, err
	}
	smtpConn = c
	return smtpConn, nil
}

func closeSMTPConnLocked() {
	if smtpConn != nil {
		_ = smtpConn.Close()
		smtpConn = nil
	}
}

func SendEmail(ctx context.Context, opt EmailOptions) error {
	_ = ctx

	from := strings.TrimSpace(opt.From)
	to := strings.TrimSpace(opt.To)
	if from == "" {
		from = deriveFromAddress()
	}
	if from == "" {
		return errors.New("missing From")
	}
	if to == "" {
		return errors.New("missing To")
	}
	if strings.TrimSpace(opt.Subject) == "" {
		return errors.New("missing Subject")
	}

	m := gomail.NewMessage()
	m.SetHeader("From", from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", opt.Subject)
	m.SetBody("text/html", opt.HTML)

	for _, a := range opt.Attachments {
		filename := strings.TrimSpace(a.Filename)
		if filename == "" {
			filename = "attachment"
		}
		contentType := strings.TrimSpace(a.ContentType)
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		m.AttachReader(filename, bytes.NewReader(a.Content), gomail.SetHeader(map[string][]string{
			"Content-Type": {contentType},
		}))
	}

	smtpMu.Lock()
	s, err := ensureSMTPConnLocked()
	smtpMu.Unlock()
	if err != nil {
		return err
	}

	sendErr := gomail.Send(s, m)
	if sendErr == nil {
		return nil
	}

	smtpMu.Lock()
	closeSMTPConnLocked()
	s, err = ensureSMTPConnLocked()
	smtpMu.Unlock()
	if err != nil {
		return sendErr
	}
	return gomail.Send(s, m)
}

func RenderDepositEmailHTML(userName string) string {
	userName = strings.TrimSpace(userName)
	if userName == "" {
		userName = "Customer"
	}
	userName = htmlEscape(userName)
	return fmt.Sprintf(`<!doctype html>
		<html>
			<body style="font-family:Arial, sans-serif; color:#111">
				<p>Hi %s,</p>
				<p>Your deposit contract is attached as PDF.</p>
				<p style="color:#666; font-size:12px">If you did not request this, please ignore.</p>
			</body>
		</html>`, userName)
}

func SendDepositEmail(ctx context.Context, contractData ContractData, attachments []AttachmentProps) error {
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	if strings.TrimSpace(contractData.UserEmail) == "" {
		return errors.New("userEmail is required")
	}
	if len(attachments) == 0 {
		return errors.New("attachments are required")
	}

	emailAtts := make([]EmailAttachment, 0, len(attachments))
	for _, a := range attachments {
		b, err := base64.StdEncoding.DecodeString(a.Content)
		if err != nil {
			return fmt.Errorf("decode attachment %q: %w", a.Filename, err)
		}
		contentType := strings.TrimSpace(a.ContentType)
		if contentType == "" {
			contentType = "application/pdf"
		}
		filename := strings.TrimSpace(a.Filename)
		if filename == "" {
			filename = "attachment.pdf"
		}
		emailAtts = append(emailAtts, EmailAttachment{
			Filename:    filename,
			Content:     b,
			ContentType: contentType,
		})
	}

	name := strings.SplitN(contractData.UserEmail, "@", 2)[0]
	html := RenderDepositEmailHTML(name)

	from := deriveFromAddress()
	if from == "" {
		return errors.New("missing From (set EMAIL_FROM or EMAIL_SERVER)")
	}

	err := SendEmail(ctx, EmailOptions{
		From:        from,
		To:          strings.TrimSpace(contractData.UserEmail),
		Subject:     fmt.Sprintf("%d - Deposit contract", contractData.NumID),
		HTML:        html,
		Attachments: emailAtts,
	})
	if err != nil {
		return err
	}

	return nil
}

func htmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
