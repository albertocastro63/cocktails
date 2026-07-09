package email

import "sync"

// StubSender records sent reset emails instead of delivering them. Used in
// tests and local/dev where SES is not configured.
type StubSender struct {
	mu   sync.Mutex
	Sent []SentReset
}

type SentReset struct {
	To   string
	Data PasswordResetData
}

func (s *StubSender) SendPasswordReset(to string, data PasswordResetData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Sent = append(s.Sent, SentReset{To: to, Data: data})
	return nil
}

// Count returns how many reset emails were recorded.
func (s *StubSender) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Sent)
}
