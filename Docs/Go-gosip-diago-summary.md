# sipgo/diago Feasibility Spike — Results Summary

Date: 2026-02-10
Phase: 0 (Feasibility Spike)
Status: **ALL TESTS PASSED — Go rewrite is greenlit**

## Overview

The Phase 0 feasibility spike validated that sipgo v1.2.0 and diago v0.27.0 can
serve as the SIP/media stack for the siptty Go rewrite. A standalone test program
(`spike/main.go`) exercised six capabilities against a real Asterisk instance
running in Docker.

## Test Results

| # | Capability | Result | diago/sipgo API used |
|---|-----------|--------|----------------------|
| 1 | Registration (digest auth) | **PASS** | `dg.RegisterTransaction()` → `t.Register()` / `t.Unregister()` |
| 2 | Outbound call (INVITE → BYE) | **PASS** | `dg.Invite()` handles 401 challenge, 100 Trying, 200 OK, ACK automatically |
| 3 | DTMF sending (RFC 4733) | **PASS** | `dialog.AudioWriterDTMF().WriteDTMF(rune)` |
| 4 | WAV file playback | **PASS** | `dialog.PlaybackCreate()` → `pb.PlayFile(path)` — 32KB written to RTP from 2s 8kHz WAV |
| 5 | Mute/unmute | **PASS** | `dialog.PlaybackControlCreate()` → `pb.Mute(true/false)` / `pb.Stop()` |
| 6 | Raw SIP message interception | **PASS** | `sip.SIPDebugTracer()` with `sip.SIPTracer` interface — full raw message capture |

## Exit Criteria Assessment

Per the design document (Section 9):

- **Items 1-2 (blockers): PASS.** Registration and outbound calls work end-to-end
  with digest authentication against Asterisk.
- **Items 3-5 (should pass): PASS.** DTMF, playback, and mute all work with
  diago's built-in APIs. No workarounds or fallbacks needed.
- **Item 6 (feasibility assessment): PASS — better than expected.** sipgo exposes
  a `SIPTracer` interface that captures every raw SIP message with direction
  (send/recv), transport protocol, local and remote addresses, and full message
  bytes. This is not a debug log scraper — it's a first-class interception hook.
  The SIP trace viewer and sngrep-style dialog viewer can be built on top of this
  without any hacks.

## Confirmed Library Versions

```
github.com/emiago/sipgo   v1.2.0
github.com/emiago/diago   v0.27.0
go                        1.23+ (diago requires >= 1.23)
```

No CGO. Pure Go. Cross-compilable with `CGO_ENABLED=0`.

## Key Technical Findings

### UA Configuration

The sipgo `UserAgent` must be configured to match Asterisk's auth expectations:

```go
ua, _ := sipgo.NewUA(
    sipgo.WithUserAgent(extension),           // e.g. "100"
    sipgo.WithUserAgentHostname("localhost"),  // hostname for From/Contact headers
)
```

Using a product name (e.g. "siptty/0.1") as the UA name breaks digest auth
because sipgo uses it to construct the From header's user part, which Asterisk
validates against the auth section's username.

### Transport and Serve Order

For call tests (INVITE + media), `ServeBackground()` must be called **before**
`RegisterTransaction()`. This ensures the transport's ephemeral port is resolved
and the client is properly configured to reuse the server's UDP socket.

```go
dg.ServeBackground(ctx, handler)  // binds port, creates client
regTx, _ := dg.RegisterTransaction(ctx, registrar, opts)  // uses resolved client
regTx.Register(ctx)
dialog, _ := dg.Invite(ctx, target, opts)  // reuses same socket
```

### Network Routing (Docker)

When Asterisk runs in Docker, the SIP signaling may go through port mapping
(host `127.0.0.1:5060` → container), but RTP uses the container's bridge IP
directly (e.g. `172.18.0.2`). Binding to `127.0.0.1` causes `sendto: invalid
argument` on RTP because Linux won't route from loopback to non-loopback
addresses.

Solution: bind to the Docker bridge gateway IP (e.g. `172.18.0.1`) or `0.0.0.0`
and target the container IP directly. In production (non-Docker), this is not an
issue — the transport binds to the real network interface.

### SIP Trace Capture Architecture

sipgo provides a clean interception point:

```go
sip.SIPDebug = true
sip.SIPDebugTracer(&myTracer{})

type myTracer struct{}
func (t *myTracer) SIPTraceRead(transport, laddr, raddr string, msg []byte)  { ... }
func (t *myTracer) SIPTraceWrite(transport, laddr, raddr string, msg []byte) { ... }
```

Each callback receives the full raw SIP message as `[]byte`, plus transport
metadata. The spike captured 66 messages across 5 call tests, including REGISTER,
INVITE, 100 Trying, 200 OK, ACK, BYE, and 401 Unauthorized — every SIP message
in both directions.

This is sufficient to build the planned `internal/trace/` package, the
`DialogTracker`, and the sngrep-style viewer without any sipgo modifications.

### Registration API

diago offers two registration patterns:

1. **`dg.Register(ctx, uri, opts)`** — blocking, handles re-registration
   automatically. Runs until context is cancelled. Good for long-lived
   registrations.
2. **`dg.RegisterTransaction(ctx, uri, opts)`** — returns a `RegisterTransaction`
   with explicit `Register()`, `Unregister()`, and `QualifyLoop()` methods. Good
   for fine-grained control.

The `RegisterOptions` struct supports `OnRegistered` callback, `Expiry`,
`RetryInterval`, and digest auth credentials.

### Mute/Unmute Implementation

diago's `PlaybackControlCreate()` returns an `AudioPlaybackControl` that wraps
`AudioPlayback` with `Mute(bool)` and `Stop()` methods. Mute works by silencing
the RTP stream (sending silence frames) rather than stopping RTP entirely. This
is the correct behavior for SIP mute — the media session stays active.

## Impact on Design Document

Based on these results:

1. **SIP trace capture can be promoted from P2 to MVP** if desired. The
   `SIPTracer` API works out of the box with zero custom sipgo work. The only
   work is the TUI panel and event wiring.
2. **All "(risk)" items in the feature mapping table are de-risked.** Mute,
   DTMF, and file audio all work as documented.
3. **No fallbacks or workarounds needed** for any tested capability.
4. **The `internal/trace/` package** can be designed with confidence that
   `SipTraceEvent` will have full raw message data with direction, transport,
   and addressing metadata.

## Spike Code

The spike is at `spike/main.go`. It is a standalone program (not part of the
test suite) that can be re-run against any Asterisk instance:

```bash
# Start Asterisk
docker compose -f docker-compose.test.yml up -d

# Run spike (auto-detects Docker bridge network)
go run ./spike/

# Or with custom addresses
SPIKE_ASTERISK_HOST=10.0.0.5:5060 SPIKE_BIND_HOST=10.0.0.1 go run ./spike/
```
