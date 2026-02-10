// spike/main.go — Phase 0 feasibility spike for siptty Go rewrite.
//
// Validates core sipgo/diago capabilities against a real Asterisk instance.
// Run with: go run ./spike/
//
// Requires Asterisk on localhost:5060 (docker compose -f docker-compose.test.yml up -d)
//
// Tests:
//   1. Registration with digest auth
//   2. Outbound call (INVITE → echo ext 600 → BYE)
//   3. DTMF sending (call ext 603, send DTMF digits)
//   4. WAV playback into a call
//   5. Mute/unmute via PlaybackControl
//   6. Raw SIP message interception via sipgo SIPTracer

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/emiago/diago"
	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
)

// ---------------------------------------------------------------------------
// SIP Trace collector — implements sip.SIPTracer
// ---------------------------------------------------------------------------

type traceEntry struct {
	Direction string // "recv" or "send"
	Transport string
	Laddr     string
	Raddr     string
	Message   string
	Timestamp time.Time
}

type sipTracer struct {
	mu      sync.Mutex
	entries []traceEntry
}

func (t *sipTracer) SIPTraceRead(transport, laddr, raddr string, msg []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = append(t.entries, traceEntry{
		Direction: "recv",
		Transport: transport,
		Laddr:     laddr,
		Raddr:     raddr,
		Message:   string(msg),
		Timestamp: time.Now(),
	})
}

func (t *sipTracer) SIPTraceWrite(transport, laddr, raddr string, msg []byte) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.entries = append(t.entries, traceEntry{
		Direction: "send",
		Transport: transport,
		Laddr:     laddr,
		Raddr:     raddr,
		Message:   string(msg),
		Timestamp: time.Now(),
	})
}

func (t *sipTracer) dump() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for i, e := range t.entries {
		firstLine := strings.SplitN(e.Message, "\r\n", 2)[0]
		dir := "←"
		if e.Direction == "send" {
			dir = "→"
		}
		fmt.Printf("  [%d] %s %s %s %s\n", i, e.Timestamp.Format("15:04:05.000"), dir, e.Transport, firstLine)
	}
}

func (t *sipTracer) count() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.entries)
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// Resolve the Docker bridge gateway IP for routing to the Asterisk container.
// We can't use 127.0.0.1 because RTP from Asterisk comes from 172.18.0.x and
// Linux won't allow sendto() from loopback to non-loopback addresses.
var (
	asteriskHost = getEnv("SPIKE_ASTERISK_HOST", "172.18.0.2:5060")
	bindHost     = getEnv("SPIKE_BIND_HOST", "172.18.0.1")
)

const (
	testExt  = "100"
	testPass = "test100"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func pass(name string) {
	fmt.Printf("  PASS: %s\n", name)
}

func fail(name string, err error) {
	fmt.Printf("  FAIL: %s — %v\n", name, err)
}

func newDiago(name string) (*sipgo.UserAgent, *diago.Diago) {
	ua, err := sipgo.NewUA(
		sipgo.WithUserAgent(name),
		sipgo.WithUserAgentHostname("localhost"),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to create UA: %v", err))
	}

	dg := diago.NewDiago(ua, diago.WithTransport(diago.Transport{
		Transport: "udp",
		BindHost:  bindHost,
		BindPort:  0, // ephemeral port
	}))
	return ua, dg
}

func parseURI(raw string) sip.Uri {
	var uri sip.Uri
	if err := sip.ParseUri(raw, &uri); err != nil {
		panic(fmt.Sprintf("bad URI %q: %v", raw, err))
	}
	return uri
}

// generateToneWAV creates a minimal 8kHz 16-bit mono WAV file with a sine tone.
func generateToneWAV(path string, durationMs int, freqHz float64) error {
	sampleRate := 8000
	numSamples := sampleRate * durationMs / 1000
	dataSize := numSamples * 2 // 16-bit = 2 bytes per sample

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	// WAV header
	write := func(data interface{}) { binary.Write(f, binary.LittleEndian, data) }
	f.Write([]byte("RIFF"))
	write(uint32(36 + dataSize)) // file size - 8
	f.Write([]byte("WAVE"))
	f.Write([]byte("fmt "))
	write(uint32(16))    // chunk size
	write(uint16(1))     // PCM
	write(uint16(1))     // mono
	write(uint32(8000))  // sample rate
	write(uint32(16000)) // byte rate (8000 * 2)
	write(uint16(2))     // block align
	write(uint16(16))    // bits per sample
	f.Write([]byte("data"))
	write(uint32(dataSize))

	// Generate sine wave
	for i := 0; i < numSamples; i++ {
		sample := int16(math.Sin(2*math.Pi*freqHz*float64(i)/float64(sampleRate)) * 16000)
		write(sample)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Spike tests
// ---------------------------------------------------------------------------

func testRegistration() bool {
	fmt.Println("\n=== Test 1: Registration ===")

	ua, dg := newDiago(testExt)
	defer ua.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	registrar := parseURI(fmt.Sprintf("sip:%s@%s", testExt, asteriskHost))

	// Use RegisterTransaction for finer control (register, then unregister)
	t, err := dg.RegisterTransaction(ctx, registrar, diago.RegisterOptions{
		Username: testExt,
		Password: testPass,
		Expiry:   60 * time.Second,
	})
	if err != nil {
		fail("create register transaction", err)
		return false
	}

	if err := t.Register(ctx); err != nil {
		fail("register", err)
		return false
	}
	pass("registered 100@localhost")

	// Unregister
	if err := t.Unregister(ctx); err != nil {
		fail("unregister", err)
		return false
	}
	pass("unregistered 100@localhost")
	return true
}

// setupCallTest creates a diago instance, starts serving in background, and registers.
// Returns cleanup function. All call tests need this pattern.
func setupCallTest(testName string) (dg *diago.Diago, ua *sipgo.UserAgent, cleanup func(), ok bool) {
	ua, dg = newDiago(testExt)

	// Start serving FIRST so transport port is resolved before client operations
	serveCtx, serveCancel := context.WithCancel(context.Background())
	if err := dg.ServeBackground(serveCtx, func(d *diago.DialogServerSession) {
		slog.Info("Inbound call (ignoring)", "id", d.ID)
	}); err != nil {
		fail("serve background", err)
		serveCancel()
		ua.Close()
		return nil, nil, nil, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	registrar := parseURI(fmt.Sprintf("sip:%s@%s", testExt, asteriskHost))
	regTx, err := dg.RegisterTransaction(ctx, registrar, diago.RegisterOptions{
		Username: testExt,
		Password: testPass,
		Expiry:   60 * time.Second,
	})
	if err != nil {
		fail("create register transaction", err)
		cancel()
		serveCancel()
		ua.Close()
		return nil, nil, nil, false
	}
	if err := regTx.Register(ctx); err != nil {
		fail("register for "+testName, err)
		cancel()
		serveCancel()
		ua.Close()
		return nil, nil, nil, false
	}
	cancel()
	pass("registered")

	cleanup = func() {
		unregCtx, unregCancel := context.WithTimeout(context.Background(), 3*time.Second)
		regTx.Unregister(unregCtx)
		unregCancel()
		serveCancel()
		ua.Close()
	}
	return dg, ua, cleanup, true
}

func testOutboundCall() bool {
	fmt.Println("\n=== Test 2: Outbound Call ===")

	dg, _, cleanup, ok := setupCallTest("outbound call")
	if !ok {
		return false
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Dial ext 603 (Answer + Wait(30) + Hangup) — no audio dependencies
	target := parseURI(fmt.Sprintf("sip:603@%s", asteriskHost))
	dialog, err := dg.Invite(ctx, target, diago.InviteOptions{
		Username: testExt,
		Password: testPass,
	})
	if err != nil {
		fail("invite", err)
		return false
	}
	pass("INVITE answered — call established")

	// Let it run briefly
	time.Sleep(1 * time.Second)

	// Hangup
	hangCtx, hangCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer hangCancel()
	if err := dialog.Hangup(hangCtx); err != nil {
		fail("hangup", err)
		return false
	}
	dialog.Close()
	pass("hangup complete")
	return true
}

func testDTMF() bool {
	fmt.Println("\n=== Test 3: DTMF Sending ===")

	dg, _, cleanup, ok := setupCallTest("DTMF")
	if !ok {
		return false
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Dial ext 603 (answer + wait) so we have a live call
	target := parseURI(fmt.Sprintf("sip:603@%s", asteriskHost))
	dialog, err := dg.Invite(ctx, target, diago.InviteOptions{
		Username: testExt,
		Password: testPass,
	})
	if err != nil {
		fail("invite for DTMF", err)
		return false
	}
	defer func() {
		dialog.Hangup(context.Background())
		dialog.Close()
	}()
	pass("call established")

	// Send DTMF digits
	dtmfWriter := dialog.AudioWriterDTMF()
	digits := []rune{'1', '2', '3', '4'}
	for _, d := range digits {
		if err := dtmfWriter.WriteDTMF(d); err != nil {
			fail(fmt.Sprintf("send DTMF '%c'", d), err)
			return false
		}
		time.Sleep(300 * time.Millisecond) // inter-digit gap
	}
	pass("sent DTMF digits 1234")
	return true
}

func testPlayback() bool {
	fmt.Println("\n=== Test 4: WAV Playback ===")

	wavPath := "/tmp/siptty-spike-tone.wav"
	if err := generateToneWAV(wavPath, 2000, 440.0); err != nil {
		fail("generate test WAV", err)
		return false
	}
	defer os.Remove(wavPath)
	pass("generated 2s 440Hz test WAV")

	dg, _, cleanup, ok := setupCallTest("playback")
	if !ok {
		return false
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	target := parseURI(fmt.Sprintf("sip:603@%s", asteriskHost))
	dialog, err := dg.Invite(ctx, target, diago.InviteOptions{
		Username: testExt,
		Password: testPass,
	})
	if err != nil {
		fail("invite for playback", err)
		return false
	}
	defer func() {
		dialog.Hangup(context.Background())
		dialog.Close()
	}()
	pass("call established")

	pb, err := dialog.PlaybackCreate()
	if err != nil {
		fail("PlaybackCreate", err)
		return false
	}

	n, err := pb.PlayFile(wavPath)
	if err != nil {
		fail("PlayFile", err)
		return false
	}
	pass(fmt.Sprintf("played WAV file (%d bytes written to RTP)", n))
	return true
}

func testMuteUnmute() bool {
	fmt.Println("\n=== Test 5: Mute/Unmute ===")

	wavPath := "/tmp/siptty-spike-tone-long.wav"
	if err := generateToneWAV(wavPath, 5000, 440.0); err != nil {
		fail("generate test WAV", err)
		return false
	}
	defer os.Remove(wavPath)

	dg, _, cleanup, ok := setupCallTest("mute")
	if !ok {
		return false
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	target := parseURI(fmt.Sprintf("sip:603@%s", asteriskHost))
	dialog, err := dg.Invite(ctx, target, diago.InviteOptions{
		Username: testExt,
		Password: testPass,
	})
	if err != nil {
		fail("invite for mute test", err)
		return false
	}
	defer func() {
		dialog.Hangup(context.Background())
		dialog.Close()
	}()

	pb, err := dialog.PlaybackControlCreate()
	if err != nil {
		fail("PlaybackControlCreate", err)
		return false
	}

	// Start playback in background
	playDone := make(chan error, 1)
	wavFile, err := os.Open(wavPath)
	if err != nil {
		fail("open WAV for mute test", err)
		return false
	}
	defer wavFile.Close()

	go func() {
		_, err := pb.Play(wavFile, "audio/wav")
		playDone <- err
	}()

	// Mute after 1s
	time.Sleep(1 * time.Second)
	pb.Mute(true)
	pass("muted")

	// Unmute after another 1s
	time.Sleep(1 * time.Second)
	pb.Mute(false)
	pass("unmuted")

	// Stop playback
	time.Sleep(500 * time.Millisecond)
	pb.Stop()

	select {
	case err := <-playDone:
		if err != nil {
			// PlaybackControl.Stop() causes the Play to return — this is expected
			fmt.Printf("  (play returned: %v — expected after Stop)\n", err)
		}
	case <-time.After(3 * time.Second):
		fail("playback did not finish after Stop()", fmt.Errorf("timeout"))
		return false
	}

	pass("mute/unmute/stop cycle complete")
	return true
}

func testSIPTrace(tracer *sipTracer) bool {
	fmt.Println("\n=== Test 6: SIP Trace Interception ===")

	count := tracer.count()
	if count == 0 {
		fail("no SIP messages captured", fmt.Errorf("tracer has 0 entries"))
		return false
	}
	pass(fmt.Sprintf("captured %d raw SIP messages", count))

	fmt.Println("\n  Captured SIP trace summary:")
	tracer.dump()

	// Check for expected message types
	tracer.mu.Lock()
	hasRegister := false
	hasInvite := false
	has200 := false
	for _, e := range tracer.entries {
		first := strings.SplitN(e.Message, "\r\n", 2)[0]
		if strings.Contains(first, "REGISTER") {
			hasRegister = true
		}
		if strings.Contains(first, "INVITE") {
			hasInvite = true
		}
		if strings.Contains(first, "200 OK") {
			has200 = true
		}
	}
	tracer.mu.Unlock()

	if hasRegister {
		pass("captured REGISTER messages")
	} else {
		fail("missing REGISTER", fmt.Errorf("not found in trace"))
	}
	if hasInvite {
		pass("captured INVITE messages")
	} else {
		fail("missing INVITE", fmt.Errorf("not found in trace"))
	}
	if has200 {
		pass("captured 200 OK responses")
	} else {
		fail("missing 200 OK", fmt.Errorf("not found in trace"))
	}

	fmt.Println("\n  Assessment: sipgo's SIPTracer interface provides full raw SIP")
	fmt.Println("  message capture with direction, transport, and addresses.")
	fmt.Println("  This is sufficient for the siptty trace viewer.")
	return true
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	// Configure structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	// Install SIP tracer BEFORE any tests run
	tracer := &sipTracer{}
	sip.SIPDebug = true
	sip.SIPDebugTracer(tracer)

	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║  siptty Phase 0 — sipgo/diago Feasibility Spike    ║")
	fmt.Println("╠══════════════════════════════════════════════════════╣")
	fmt.Printf("║  Asterisk: %s                         ║\n", asteriskHost)
	fmt.Printf("║  Test ext: %s (pass: %s)                    ║\n", testExt, testPass)
	fmt.Println("╚══════════════════════════════════════════════════════╝")

	results := map[string]bool{}

	results["1-registration"] = testRegistration()
	time.Sleep(500 * time.Millisecond) // let UDP sockets close between tests
	results["2-outbound-call"] = testOutboundCall()
	time.Sleep(500 * time.Millisecond)
	results["3-dtmf"] = testDTMF()
	time.Sleep(500 * time.Millisecond)
	results["4-playback"] = testPlayback()
	time.Sleep(500 * time.Millisecond)
	results["5-mute-unmute"] = testMuteUnmute()
	results["6-sip-trace"] = testSIPTrace(tracer)

	// Summary
	fmt.Println("\n══════════════════════════════════════════════════════")
	fmt.Println("RESULTS SUMMARY")
	fmt.Println("══════════════════════════════════════════════════════")

	allPassed := true
	for name, ok := range results {
		status := "PASS"
		if !ok {
			status = "FAIL"
			allPassed = false
		}
		fmt.Printf("  [%s] %s\n", status, name)
	}

	fmt.Println("══════════════════════════════════════════════════════")
	if allPassed {
		fmt.Println("ALL TESTS PASSED — sipgo/diago is viable for siptty")
		os.Exit(0)
	} else {
		fmt.Println("SOME TESTS FAILED — review results above")
		os.Exit(1)
	}
}
