package utils

// import (
// 	"bytes"
// 	"context"
// 	"crypto/tls"
// 	"encoding/base64"
// 	"errors"
// 	"fmt"
// 	"net/url"
// 	"os"
// 	"strconv"
// 	"strings"
// 	"sync"
// 	"time"

// 	gomail "gopkg.in/mail.v2"
// )

// type EmailAttachment struct {
// 	Filename    string
// 	Content     []byte
// 	ContentType string
// }

// type EmailOptions struct {
// 	From        string
// 	To          string
// 	Subject     string
// 	HTML        string
// 	Attachments []EmailAttachment
// }

// var (
// 	smtpMu     sync.Mutex
// 	smtpDialer *gomail.Dialer
// 	smtpConn   gomail.SendCloser
// )

// func smtpTimeout() time.Duration {
// 	if v := strings.TrimSpace(os.Getenv("EMAIL_TIMEOUT_SECONDS")); v != "" {
// 		if n, err := strconv.Atoi(v); err == nil && n > 0 {
// 			return time.Duration(n) * time.Second
// 		}
// 	}
// 	return 20 * time.Second
// }

// func tlsConfigForHost(host string) *tls.Config {
// 	return &tls.Config{
// 		ServerName:         host,
// 		MinVersion:         tls.VersionTLS12,
// 		InsecureSkipVerify: strings.EqualFold(strings.TrimSpace(os.Getenv("EMAIL_TLS_INSECURE")), "true"),
// 	}
// }
