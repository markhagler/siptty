package engine

import (
	"sync"
	"time"

	"github.com/emiago/diago"
)

// Call wraps a diago dialog session with metadata for the TUI.
type Call struct {
	ID        string
	Direction string // "inbound", "outbound"
	RemoteURI string
	State     string // "calling", "incoming", "early", "confirmed", "disconnected"
	StartTime time.Time

	mu       sync.Mutex
	client   *diago.DialogClientSession
	server   *diago.DialogServerSession
	answerCh chan struct{} // signals the inbound handler to accept
}

// newOutboundCall creates a Call for an outgoing INVITE.
func newOutboundCall(id, remoteURI string, client *diago.DialogClientSession) *Call {
	return &Call{
		ID:        id,
		Direction: "outbound",
		RemoteURI: remoteURI,
		State:     "calling",
		StartTime: time.Now(),
		client:    client,
	}
}

// newInboundCall creates a Call for an incoming INVITE.
func newInboundCall(id, remoteURI string, server *diago.DialogServerSession) *Call {
	return &Call{
		ID:        id,
		Direction: "inbound",
		RemoteURI: remoteURI,
		State:     "incoming",
		StartTime: time.Now(),
		server:    server,
		answerCh:  make(chan struct{}, 1),
	}
}

// setState updates the call state under lock.
func (c *Call) setState(state string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.State = state
}

// ValidDTMFDigit returns true if r is a valid DTMF digit (0-9, *, #, A-D).
func ValidDTMFDigit(r rune) bool {
	switch {
	case r >= '0' && r <= '9':
		return true
	case r == '*' || r == '#':
		return true
	case r >= 'A' && r <= 'D':
		return true
	}
	return false
}
