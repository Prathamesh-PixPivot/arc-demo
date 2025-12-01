package services

import (
	"errors"
	"fmt"
	"io"
	"time"

	"gopkg.in/gomail.v2"
)

// Email error types
var (
	ErrSMTPConnection = errors.New("SMTP connection failed")
	ErrSMTPAuth       = errors.New("SMTP authentication failed")
	ErrEmailSend      = errors.New("failed to send email")
)

type EmailService struct {
	dialer *gomail.Dialer
	from   string
}

func NewEmailService(host string, port int, username, password, from string) *EmailService {
	dialer := gomail.NewDialer(host, port, username, password)
	return &EmailService{dialer: dialer, from: from}
}

// TestConnection validates SMTP configuration by attempting to connect
func (s *EmailService) TestConnection() error {
	d, err := s.dialer.Dial()
	if err != nil {
		if isAuthError(err) {
			return fmt.Errorf("%w: %v", ErrSMTPAuth, err)
		}
		return fmt.Errorf("%w: %v", ErrSMTPConnection, err)
	}
	defer d.Close()
	return nil
}

// Send sends an email with retry logic (up to 3 attempts)
func (s *EmailService) Send(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	return s.sendWithRetry(m, 3)
}

func (s *EmailService) SendEmailWithAttachment(to, subject, body string, attachment []byte, filename string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	// Add attachment using SetCopyFunc
	m.Attach(filename, gomail.SetCopyFunc(func(w io.Writer) error {
		_, err := w.Write(attachment)
		return err
	}))

	return s.sendWithRetry(m, 3)
}

// sendWithRetry attempts to send email with exponential backoff
func (s *EmailService) sendWithRetry(m *gomail.Message, maxAttempts int) error {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := s.dialer.DialAndSend(m)
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Don't retry on authentication errors
		if isAuthError(err) {
			return fmt.Errorf("%w: %v", ErrSMTPAuth, err)
		}

		// If not the last attempt, wait before retrying
		if attempt < maxAttempts {
			// Exponential backoff: 1s, 2s, 4s
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("%w after %d attempts: %v", ErrEmailSend, maxAttempts, lastErr)
}

// isAuthError checks if the error is authentication-related
func isAuthError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	// Common SMTP auth error patterns
	return contains(errStr, "535") || // Authentication failed
		contains(errStr, "535 5.7.8") || // Bad credentials
		contains(errStr, "auth") ||
		contains(errStr, "authentication")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(hasPrefix(s, substr) || hasSuffix(s, substr) || containsSubstr(s, substr)))
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
