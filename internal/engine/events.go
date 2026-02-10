package engine

import "time"

// Event is the interface for all engine-to-TUI events.
type Event interface {
	eventMarker()
}

// RegStateEvent reports registration state changes for an account.
type RegStateEvent struct {
	AccountID string
	State     string // "registered", "unregistered", "failed"
	Reason    string
}

func (RegStateEvent) eventMarker() {}

// CallStateEvent reports call state transitions.
type CallStateEvent struct {
	CallID    string
	State     string // "calling", "incoming", "early", "confirmed", "disconnected"
	RemoteURI string
	Duration  time.Duration
	Direction string // "inbound", "outbound"
}

func (CallStateEvent) eventMarker() {}

// SipTraceEvent carries a raw SIP message captured by the sipgo SIPTracer.
type SipTraceEvent struct {
	Direction  string    // "send", "recv"
	Message    string    // full raw SIP message text
	Timestamp  time.Time
	Transport  string    // "udp", "tcp", "tls"
	LocalAddr  string
	RemoteAddr string
}

func (SipTraceEvent) eventMarker() {}

// DTMFEvent reports a received DTMF digit on a call.
type DTMFEvent struct {
	CallID string
	Digit  rune
}

func (DTMFEvent) eventMarker() {}
