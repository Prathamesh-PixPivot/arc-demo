package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
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
	dialer   *gomail.Dialer
	from     string
	mockMode bool // ✅ ADDED: Flag to enable local mocking
}

func NewEmailService(host string, port int, username, password, from string) *EmailService {
	// ✅ ADDED: Auto-detect if we should run in Mock Mode
	if host == "" || username == "" || password == "" {
		log.Println("[EmailService] ⚠️ SMTP credentials missing. Running in MOCK MODE (Emails will be logged to stdout)")
		return &EmailService{
			dialer:   nil,
			from:     from,
			mockMode: true,
		}
	}

	dialer := gomail.NewDialer(host, port, username, password)
	return &EmailService{dialer: dialer, from: from, mockMode: false}
}

// TestConnection validates SMTP configuration by attempting to connect
func (s *EmailService) TestConnection() error {
	// ✅ ADDED: Mock Mode check
	if s.mockMode {
		log.Println("[EmailService] TestConnection: Skipping connection check (Mock Mode Active)")
		return nil
	}

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
	// ✅ ADDED: Mock Mode Handler
	if s.mockMode {
		log.Printf("📧 [MOCK EMAIL] To: %s | Subject: %s\nBody: %s\n---------------------------------------------------", to, subject, body)
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	return s.sendWithRetry(m, 3)
}

func (s *EmailService) SendEmailWithAttachment(to, subject, body string, attachment []byte, filename string) error {
	// ✅ ADDED: Mock Mode Handler
	if s.mockMode {
		log.Printf("📧 [MOCK EMAIL + ATTACHMENT] To: %s | Subject: %s | File: %s\n---------------------------------------------------", to, subject, filename)
		return nil
	}

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
	return strings.Contains(errStr, "535") || // Authentication failed
		strings.Contains(errStr, "535 5.7.8") || // Bad credentials
		strings.Contains(errStr, "auth") ||
		strings.Contains(errStr, "authentication")
}
