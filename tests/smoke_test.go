//go:build integration

package tests

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/siptty/siptty/internal/config"
	"github.com/siptty/siptty/internal/engine"
)

// Smoke test: verifies the MVP engine against a live Asterisk instance.
// Run with: go test -tags integration -v ./tests/
//
// Requires Asterisk running (docker compose -f docker-compose.test.yml up -d)

func getTestEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestSmokeRegistration(t *testing.T) {
	asteriskHost := getTestEnv("SMOKE_ASTERISK_HOST", "172.18.0.2")
	bindHost := getTestEnv("SMOKE_BIND_HOST", "172.18.0.1")

	cfg := &config.Config{
		General: config.GeneralConfig{
			LogLevel:  4,
			UserAgent: "siptty-smoke",
			BindHost:  bindHost,
		},
		Accounts: []config.AccountConfig{
			{
				Name:         "smoke-100",
				Enabled:      true,
				SipURI:       fmt.Sprintf("sip:100@%s", asteriskHost),
				AuthUser:     "100",
				AuthPassword: "test100",
				Registrar:    fmt.Sprintf("sip:100@%s:5060", asteriskHost),
				Transport:    "udp",
				Register:     true,
				RegExpiry:    60,
			},
		},
		Audio: config.AudioConfig{Mode: "null"},
	}

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := eng.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer eng.Stop()

	// Collect events for a few seconds.
	var regEvents []engine.RegStateEvent
	var traceEvents []engine.SipTraceEvent

	deadline := time.After(10 * time.Second)
	registered := false

loop:
	for {
		select {
		case ev, ok := <-eng.Events():
			if !ok {
				break loop
			}
			switch e := ev.(type) {
			case engine.RegStateEvent:
				regEvents = append(regEvents, e)
				t.Logf("REG EVENT: account=%s state=%s reason=%s", e.AccountID, e.State, e.Reason)
				if e.State == "registered" {
					registered = true
					// Give a moment for more trace events to arrive.
					time.Sleep(500 * time.Millisecond)
					break loop
				}
			case engine.SipTraceEvent:
				traceEvents = append(traceEvents, e)
				firstLine := strings.SplitN(e.Message, "\r\n", 2)[0]
				t.Logf("SIP TRACE: %s %s %s", e.Direction, e.Transport, firstLine)
			}
		case <-deadline:
			break loop
		}
	}

	// Verify registration succeeded.
	if !registered {
		t.Errorf("registration did not succeed within timeout; got %d reg events", len(regEvents))
		for _, e := range regEvents {
			t.Logf("  reg event: state=%s reason=%s", e.State, e.Reason)
		}
	}

	// Verify SIP trace events were captured.
	if len(traceEvents) == 0 {
		t.Error("no SIP trace events captured")
	} else {
		t.Logf("captured %d SIP trace events", len(traceEvents))

		hasRegister := false
		has200 := false
		has401 := false
		for _, e := range traceEvents {
			first := strings.SplitN(e.Message, "\r\n", 2)[0]
			if strings.Contains(first, "REGISTER") {
				hasRegister = true
			}
			if strings.Contains(first, "200 OK") {
				has200 = true
			}
			if strings.Contains(first, "401") {
				has401 = true
			}
		}
		if !hasRegister {
			t.Error("no REGISTER messages in trace")
		}
		if !has401 {
			t.Log("note: no 401 challenge seen (may have been cached)")
		}
		if !has200 {
			t.Error("no 200 OK in trace")
		}
	}
}

func TestSmokeOutboundCall(t *testing.T) {
	asteriskHost := getTestEnv("SMOKE_ASTERISK_HOST", "172.18.0.2")
	bindHost := getTestEnv("SMOKE_BIND_HOST", "172.18.0.1")

	cfg := &config.Config{
		General: config.GeneralConfig{
			LogLevel:  4,
			UserAgent: "siptty-smoke",
			BindHost:  bindHost,
		},
		Accounts: []config.AccountConfig{
			{
				Name:         "smoke-100",
				Enabled:      true,
				SipURI:       fmt.Sprintf("sip:100@%s", asteriskHost),
				AuthUser:     "100",
				AuthPassword: "test100",
				Registrar:    fmt.Sprintf("sip:100@%s:5060", asteriskHost),
				Transport:    "udp",
				Register:     true,
				RegExpiry:    60,
			},
		},
		Audio: config.AudioConfig{Mode: "null"},
	}

	eng, err := engine.NewEngine(cfg)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := eng.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer eng.Stop()

	// Wait for registration.
	deadline := time.After(10 * time.Second)
	registered := false
regLoop:
	for {
		select {
		case ev, ok := <-eng.Events():
			if !ok {
				break regLoop
			}
			if e, ok := ev.(engine.RegStateEvent); ok {
				t.Logf("REG: %s %s", e.AccountID, e.State)
				if e.State == "registered" {
					registered = true
					break regLoop
				}
			}
		case <-deadline:
			break regLoop
		}
	}
	if !registered {
		t.Fatal("registration failed — cannot test outbound call")
	}

	// Dial ext 603 (Answer + Wait(30)).
	target := fmt.Sprintf("sip:603@%s:5060", asteriskHost)
	if err := eng.Dial("smoke-100", target); err != nil {
		t.Fatalf("Dial: %v", err)
	}

	// Wait for call to be confirmed.
	callDeadline := time.After(15 * time.Second)
	confirmed := false
	var callID string
	var traceCount int
callLoop:
	for {
		select {
		case ev, ok := <-eng.Events():
			if !ok {
				break callLoop
			}
			switch e := ev.(type) {
			case engine.CallStateEvent:
				t.Logf("CALL: id=%s state=%s remote=%s", e.CallID, e.State, e.RemoteURI)
				callID = e.CallID
				if e.State == "confirmed" {
					confirmed = true
					break callLoop
				}
				if e.State == "disconnected" {
					t.Logf("call disconnected before confirmed")
					break callLoop
				}
			case engine.SipTraceEvent:
				traceCount++
			}
		case <-callDeadline:
			break callLoop
		}
	}

	if !confirmed {
		t.Errorf("outbound call was not confirmed")
	} else {
		t.Logf("call %s confirmed — SIP traces during call: %d", callID, traceCount)

		// Hangup.
		if err := eng.Hangup(callID); err != nil {
			t.Errorf("Hangup: %v", err)
		}
		// Drain the disconnect event.
		time.Sleep(500 * time.Millisecond)
	}
}
