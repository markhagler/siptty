package engine

import (
	"testing"
	"time"
)

func TestValidDTMFDigit(t *testing.T) {
	valid := []rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '*', '#', 'A', 'B', 'C', 'D'}
	for _, r := range valid {
		if !ValidDTMFDigit(r) {
			t.Errorf("expected %c to be valid DTMF digit", r)
		}
	}

	invalid := []rune{'a', 'b', 'E', 'z', '!', '@', ' ', 'F', 'G'}
	for _, r := range invalid {
		if ValidDTMFDigit(r) {
			t.Errorf("expected %c to be invalid DTMF digit", r)
		}
	}
}

func TestEventTypes(t *testing.T) {
	// Verify all event types satisfy the Event interface.
	var events []Event

	events = append(events, RegStateEvent{
		AccountID: "alice",
		State:     "registered",
	})
	events = append(events, CallStateEvent{
		CallID:    "1",
		State:     "confirmed",
		RemoteURI: "sip:100@pbx.io",
		Duration:  5 * time.Second,
		Direction: "outbound",
	})
	events = append(events, SipTraceEvent{
		Direction:  "send",
		Message:    "INVITE sip:100@pbx.io SIP/2.0\r\n",
		Timestamp:  time.Now(),
		Transport:  "udp",
		LocalAddr:  "10.0.0.1:5060",
		RemoteAddr: "10.0.0.2:5060",
	})
	events = append(events, DTMFEvent{
		CallID: "1",
		Digit:  '5',
	})

	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
}

func TestCallState(t *testing.T) {
	call := &Call{
		ID:        "1",
		Direction: "outbound",
		RemoteURI: "sip:100@pbx.io",
		State:     "calling",
		StartTime: time.Now(),
	}

	call.setState("confirmed")
	if call.State != "confirmed" {
		t.Errorf("expected state 'confirmed', got %q", call.State)
	}

	call.setState("disconnected")
	if call.State != "disconnected" {
		t.Errorf("expected state 'disconnected', got %q", call.State)
	}
}
