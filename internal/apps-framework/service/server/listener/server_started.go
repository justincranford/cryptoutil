// Copyright (c) 2025-2026 Justin Cranford.
// Package listener provides a reusable template for dual HTTPS server pattern used across all cryptoutil services.
package listener

// Started returns a channel closed after the public listener is bound.
func (s *PublicHTTPServer) Started() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.startedCh == nil {
		s.startedCh = make(chan struct{})
	}

	return s.startedCh
}

func (s *PublicHTTPServer) signalStarted() {
	if !s.started {
		if s.startedCh == nil {
			s.startedCh = make(chan struct{})
		}

		close(s.startedCh)
		s.started = true
	}
}
