package engine

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/emiago/diago"
	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/siptty/siptty/internal/config"
)

// Engine owns the diago instance and provides a clean API to the TUI.
type Engine struct {
	dg     *diago.Diago
	ua     *sipgo.UserAgent
	config *config.Config
	events chan Event

	accounts map[string]*Account
	calls    map[string]*Call
	mu       sync.RWMutex

	serveCancel context.CancelFunc
	nextCallID  int
}

// sipTracer implements sipgo's sip.SIPTracer interface to capture raw SIP messages.
type sipTracer struct {
	events chan<- Event
	once   sync.Once
}

func (t *sipTracer) SIPTraceRead(transport, laddr, raddr string, msg []byte) {
	ev := SipTraceEvent{
		Direction:  "recv",
		Message:    string(msg),
		Timestamp:  time.Now(),
		Transport:  transport,
		LocalAddr:  laddr,
		RemoteAddr: raddr,
	}
	select {
	case t.events <- ev:
	default:
		t.dropWarn()
	}
}

func (t *sipTracer) SIPTraceWrite(transport, laddr, raddr string, msg []byte) {
	ev := SipTraceEvent{
		Direction:  "send",
		Message:    string(msg),
		Timestamp:  time.Now(),
		Transport:  transport,
		LocalAddr:  laddr,
		RemoteAddr: raddr,
	}
	select {
	case t.events <- ev:
	default:
		t.dropWarn()
	}
}

func (t *sipTracer) dropWarn() {
	t.once.Do(func() {
		slog.Warn("SIP trace events dropped: channel full, TUI too slow")
	})
}

// NewEngine creates a new engine from the config.
// The sipgo UA name is set to the first account's extension (user part of SIP URI)
// because Asterisk validates digest auth against the From header user part.
func NewEngine(cfg *config.Config) (*Engine, error) {
	e := &Engine{
		config:   cfg,
		events:   make(chan Event, 256),
		accounts: make(map[string]*Account),
		calls:    make(map[string]*Call),
	}

	// Install SIP tracer before creating UA so all messages are captured.
	sip.SIPDebug = true
	sip.SIPDebugTracer(&sipTracer{events: e.events})

	// UA name must be the SIP extension for digest auth to work with Asterisk.
	uaName := deriveExtension(cfg)
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent(uaName),
		sipgo.WithUserAgentHostname("localhost"),
	)
	if err != nil {
		return nil, fmt.Errorf("creating sipgo UA: %w", err)
	}
	e.ua = ua

	// Configure diago transport from config.
	transport := "udp"
	if len(cfg.Accounts) > 0 {
		transport = cfg.Accounts[0].Transport
	}
	e.dg = diago.NewDiago(ua, diago.WithTransport(diago.Transport{
		Transport: transport,
		BindHost:  cfg.General.BindHost,
		BindPort:  cfg.General.BindPort,
	}))

	// Set up account structs.
	for _, acctCfg := range cfg.Accounts {
		if !acctCfg.Enabled {
			continue
		}
		a := &Account{
			ID:     acctCfg.Name,
			Config: acctCfg,
			State:  "unregistered",
		}
		e.accounts[acctCfg.Name] = a
	}

	return e, nil
}

// Events returns the read-only event channel for the TUI.
func (e *Engine) Events() <-chan Event {
	return e.events
}

// Accounts returns the IDs of all configured accounts.
func (e *Engine) Accounts() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	ids := make([]string, 0, len(e.accounts))
	for id := range e.accounts {
		ids = append(ids, id)
	}
	return ids
}

// Start begins serving (for inbound calls) and registers all accounts.
// ServeBackground must be called BEFORE RegisterTransaction (spike lesson).
func (e *Engine) Start(ctx context.Context) error {
	serveCtx, serveCancel := context.WithCancel(ctx)
	e.serveCancel = serveCancel

	if err := e.dg.ServeBackground(serveCtx, e.inboundHandler); err != nil {
		serveCancel()
		return fmt.Errorf("serve background: %w", err)
	}

	// Register all enabled accounts in goroutines.
	for _, acct := range e.accounts {
		if acct.Config.Register {
			go acct.register(ctx, e.dg, e.events)
		}
	}

	return nil
}

// Stop unregisters all accounts and shuts down diago.
func (e *Engine) Stop() {
	for _, acct := range e.accounts {
		acct.unregister()
	}
	if e.serveCancel != nil {
		e.serveCancel()
	}
	if e.ua != nil {
		e.ua.Close()
	}
	close(e.events)
}

// Dial initiates an outbound call from the specified account to the target URI.
func (e *Engine) Dial(accountID, uri string) error {
	e.mu.RLock()
	acct, ok := e.accounts[accountID]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("account %q not found", accountID)
	}

	go e.dialAsync(acct, uri)
	return nil
}

func (e *Engine) dialAsync(acct *Account, uri string) {
	var target sip.Uri
	if err := sip.ParseUri(uri, &target); err != nil {
		slog.Error("invalid dial URI", "uri", uri, "error", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	e.mu.Lock()
	e.nextCallID++
	callID := fmt.Sprintf("%d", e.nextCallID)
	e.mu.Unlock()

	e.events <- CallStateEvent{
		CallID:    callID,
		State:     "calling",
		RemoteURI: uri,
		Direction: "outbound",
	}

	dialog, err := e.dg.Invite(ctx, target, diago.InviteOptions{
		Username: acct.Config.AuthUser,
		Password: acct.Config.AuthPassword,
	})
	if err != nil {
		slog.Error("invite failed", "uri", uri, "error", err)
		e.events <- CallStateEvent{
			CallID:    callID,
			State:     "disconnected",
			RemoteURI: uri,
			Direction: "outbound",
		}
		return
	}

	call := newOutboundCall(callID, uri, dialog)
	call.setState("confirmed")

	e.mu.Lock()
	e.calls[callID] = call
	e.mu.Unlock()

	e.events <- CallStateEvent{
		CallID:    callID,
		State:     "confirmed",
		RemoteURI: uri,
		Direction: "outbound",
	}
}

// Answer accepts an incoming call.
func (e *Engine) Answer(callID string) error {
	e.mu.RLock()
	call, ok := e.calls[callID]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("call %q not found", callID)
	}
	if call.answerCh == nil {
		return fmt.Errorf("call %q is not an inbound call", callID)
	}

	select {
	case call.answerCh <- struct{}{}:
	default:
	}
	return nil
}

// Hangup terminates a call.
func (e *Engine) Hangup(callID string) error {
	e.mu.RLock()
	call, ok := e.calls[callID]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("call %q not found", callID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var err error
	if call.client != nil {
		err = call.client.Hangup(ctx)
	} else if call.server != nil {
		err = call.server.Hangup(ctx)
	}

	call.setState("disconnected")
	e.events <- CallStateEvent{
		CallID:    callID,
		State:     "disconnected",
		RemoteURI: call.RemoteURI,
		Direction: call.Direction,
	}

	return err
}

// SendDTMF sends a DTMF digit on the specified call.
func (e *Engine) SendDTMF(callID string, digit rune) error {
	if !ValidDTMFDigit(digit) {
		return fmt.Errorf("invalid DTMF digit: %c", digit)
	}

	e.mu.RLock()
	call, ok := e.calls[callID]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("call %q not found", callID)
	}

	if call.client != nil {
		return call.client.AudioWriterDTMF().WriteDTMF(digit)
	}
	if call.server != nil {
		return call.server.AudioWriterDTMF().WriteDTMF(digit)
	}
	return fmt.Errorf("call %q has no active dialog", callID)
}

// Transfer performs a blind transfer (REFER) of the specified call.
func (e *Engine) Transfer(callID, target string) error {
	e.mu.RLock()
	call, ok := e.calls[callID]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("call %q not found", callID)
	}

	var targetURI sip.Uri
	if err := sip.ParseUri(target, &targetURI); err != nil {
		return fmt.Errorf("invalid transfer target %q: %w", target, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if call.client != nil {
		return call.client.Refer(ctx, targetURI)
	}
	if call.server != nil {
		return call.server.Refer(ctx, targetURI)
	}
	return fmt.Errorf("call %q has no active dialog", callID)
}

// PlayAudio plays a WAV file into the specified call.
func (e *Engine) PlayAudio(callID, path string) error {
	e.mu.RLock()
	call, ok := e.calls[callID]
	e.mu.RUnlock()
	if !ok {
		return fmt.Errorf("call %q not found", callID)
	}

	if call.client != nil {
		pb, err := call.client.PlaybackCreate()
		if err != nil {
			return fmt.Errorf("playback create: %w", err)
		}
		_, err = pb.PlayFile(path)
		return err
	}
	if call.server != nil {
		pb, err := call.server.PlaybackCreate()
		if err != nil {
			return fmt.Errorf("playback create: %w", err)
		}
		_, err = pb.PlayFile(path)
		return err
	}
	return fmt.Errorf("call %q has no active dialog", callID)
}

// inboundHandler is called by diago for each incoming INVITE.
func (e *Engine) inboundHandler(d *diago.DialogServerSession) {
	remoteURI := d.InviteRequest.From().Address.String()

	e.mu.Lock()
	e.nextCallID++
	callID := fmt.Sprintf("%d", e.nextCallID)
	call := newInboundCall(callID, remoteURI, d)
	e.calls[callID] = call
	e.mu.Unlock()

	slog.Info("incoming call", "id", callID, "from", remoteURI)

	e.events <- CallStateEvent{
		CallID:    callID,
		State:     "incoming",
		RemoteURI: remoteURI,
		Direction: "inbound",
	}

	// Wait for answer signal or context cancellation.
	ctx := d.Context()
	select {
	case <-call.answerCh:
		if err := d.Answer(); err != nil {
			slog.Error("answer failed", "id", callID, "error", err)
			call.setState("disconnected")
			e.events <- CallStateEvent{
				CallID:    callID,
				State:     "disconnected",
				RemoteURI: remoteURI,
				Direction: "inbound",
			}
			return
		}
		call.setState("confirmed")
		e.events <- CallStateEvent{
			CallID:    callID,
			State:     "confirmed",
			RemoteURI: remoteURI,
			Direction: "inbound",
		}

		// Block until call ends.
		<-ctx.Done()

	case <-ctx.Done():
		// Caller hung up or timeout before we answered.
	}

	call.setState("disconnected")
	e.events <- CallStateEvent{
		CallID:    callID,
		State:     "disconnected",
		RemoteURI: remoteURI,
		Direction: "inbound",
	}
}

// deriveExtension returns the user part of the first account's SIP URI.
// sipgo uses this as the From header user, which Asterisk checks for digest auth.
func deriveExtension(cfg *config.Config) string {
	if len(cfg.Accounts) == 0 {
		return "siptty"
	}
	uri := cfg.Accounts[0].SipURI
	uri = strings.TrimPrefix(uri, "sips:")
	uri = strings.TrimPrefix(uri, "sip:")
	if idx := strings.Index(uri, "@"); idx >= 0 {
		return uri[:idx]
	}
	return uri
}
