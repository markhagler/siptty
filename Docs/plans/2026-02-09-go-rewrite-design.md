# siptty Go Rewrite — Design Document

Date: 2026-02-09
Revised: 2026-02-10 — Promoted SIP trace capture to Phase 1 MVP; demoted
mute/unmute to Phase 2 (not needed with null/canned audio). Phase 0 spike
completed; all 6 validations passed; de-risked all "(risk)" items.
Previous: 2026-02-10 — Phase 0 spike completed; SIP trace capture validated
via sipgo SIPTracer API.
Previous: 2026-02-09 — Incorporated reviewer feedback; added Phase 0 spike;
deferred SIP trace/dialog viewer; redesigned TUI for multi-call and BLF.

## 1. Goals & Distribution

siptty is a terminal SIP softphone for VoIP engineers, rebuilt in Go for
single-binary distribution. A VoIP engineer downloads one file, runs it,
and has a full interactive phone for testing, debugging, and emulating SIP
client devices.

### Distribution model

```
GOOS=linux   GOARCH=amd64  go build -o siptty-linux-amd64
GOOS=linux   GOARCH=arm64  go build -o siptty-linux-arm64
GOOS=darwin  GOARCH=arm64  go build -o siptty-darwin-arm64
GOOS=darwin  GOARCH=amd64  go build -o siptty-darwin-amd64
GOOS=windows GOARCH=amd64  go build -o siptty-windows-amd64.exe
```

No Python, no pip, no venv, no SWIG, no Docker, no shared libraries.
`CGO_ENABLED=0`. Fully static binary, ~10-20 MB.

### Tech stack

| Layer | Choice | Why |
|-------|--------|-----|
| SIP signaling | sipgo (BSD-2) | Mature Go SIP stack, v1.0 stable API |
| Call control + RTP | diago (MPL-2.0) | High-level dialog/media on top of sipgo |
| TUI | tview + tcell | Grid layouts, tables, text views — maps to existing design |
| Config | TOML (BurntSushi/toml) | Same format as Python version, human-friendly |
| Codecs | G.711 ulaw/alaw (pure Go) | Universal compatibility, no CGO |
| Build | `go build` | Cross-compile from any host |
| Testing | Go stdlib `testing` + testcontainers-go | Asterisk in Docker for integration tests |

### What we keep from the Python design

The TOML config format, the TUI layout, the feature tiers, the Asterisk test
environment.

### What we drop

PJSIP, SWIG, Docker build stages, PyInstaller, the entire native-library
dependency chain.

---

## 2. Architecture

The Go version follows the same layered architecture as the Python design — TUI
layer on top, engine layer in the middle, SIP stack at the bottom — but the
threading model is simpler because Go's goroutines and channels replace the
Python `call_from_thread()` bridge.

```
┌─────────────────────────────────────────────────────┐
│                 tview TUI (main goroutine)           │
│  ┌────────────┐ ┌──────────────┐ ┌──────────────┐   │
│  │ Accounts   │ │ Call Control  │ │ BLF Panel    │   │
│  │ Panel      │ │ & Dialpad    │ │ (placeholder) │   │
│  └────────────┘ └──────────────┘ └──────────────┘   │
│  ┌─────────────────────────────────────────────┐     │
│  │    Tabbed: SIP Trace / Calls / Dialogs      │     │
│  └─────────────────────────────────────────────┘     │
└──────────────────────┬──────────────────────────────┘
                       │ events via channel
                       │ commands via method calls
┌──────────────────────┼──────────────────────────────┐
│              SIP Engine (owns diago)                  │
│  ┌────────────┐ ┌──────────────┐ ┌──────────────┐   │
│  │ Account    │ │ Call         │ │ Header       │   │
│  │ Manager    │ │ Manager      │ │ Override     │   │
│  └────────────┘ └──────────────┘ └──────────────┘   │
└──────────────────────┬──────────────────────────────┘
                       │
              diago (sipgo + RTP)
                       │
                UDP/TCP/TLS + RTP
```

### Communication pattern

- **TUI to Engine:** Direct method calls. tview key bindings call engine
  methods like `engine.Dial()`, `engine.Answer()`, `engine.Hangup()`. These
  run in goroutines to avoid blocking the TUI.
- **Engine to TUI:** A single `chan EngineEvent` channel. The engine pushes
  typed events (registration state changes, call state changes, SIP trace
  messages). The TUI reads this channel and calls `tview.App.QueueUpdateDraw()`
  to safely update widgets from the event-processing goroutine.
- **diago to Engine:** diago's `ServeBackground()` runs in its own goroutine.
  Incoming calls arrive as handler callbacks, each in their own goroutine. The
  engine wraps these into events and pushes them onto the channel.

### Forward-compatibility note

The event channel and engine struct are designed to accept additional event types
without structural changes. `SipTraceEvent` is wired end-to-end in MVP — the
engine emits trace events via sipgo's `SIPTracer` interface and the TUI SIP
Trace tab displays them in real time. The SIP Dialogs tab remains a placeholder
for the Phase 2 sngrep-style dialog viewer, which will consume the same trace
events through a `DialogTracker`.

---

## 3. Engine Layer & diago Integration

The engine is a Go struct that owns the diago instance and exposes a clean API
to the TUI.

### Engine struct

```go
type Engine struct {
    dg        *diago.Diago
    config    *Config
    events    chan Event        // Engine → TUI
    accounts  map[string]*Account
    calls     map[string]*Call  // keyed by dialog ID
    mu        sync.RWMutex
}
```

### Account lifecycle

On startup, the engine loops through configured accounts and calls
`dg.Register()` for each in a goroutine. diago handles digest auth, refresh
timers, and NAT detection automatically. The `OnRegistered` callback pushes a
`RegStateEvent` onto the channel. Unregistration is handled by canceling the
registration context.

### Outbound calls

`engine.Dial(accountID, uri)` calls `dg.Invite()` in a goroutine. diago
handles the full INVITE transaction — SDP offer/answer, codec negotiation, RTP
session setup, ACK. The engine wraps the resulting `DialogClientSession` in a
`Call` struct, stores it, and pushes `CallStateEvent` updates as the dialog
progresses.

### Inbound calls

`dg.ServeBackground()` invokes the handler for each incoming INVITE. The
handler pushes a `CallStateEvent{State: "incoming"}` onto the channel. The TUI
displays the incoming call. When the user presses `a`, the TUI calls
`engine.Answer(callID)` which calls `dialog.Answer()` inside the handler
goroutine (coordinated via a channel on the `Call` struct).

### Call operations mapping to diago

| Operation | Engine method | diago call |
|-----------|--------------|------------|
| Dial | `Dial(account, uri)` | `dg.Invite(ctx, uri, opts)` |
| Answer | `Answer(callID)` | `dialog.Answer()` |
| Hangup | `Hangup(callID)` | `dialog.Hangup(ctx)` |
| Send DTMF | `SendDTMF(callID, digits)` | `dialog.AudioWriterDTMF().WriteDTMF(r)` |
| Mute | `Mute(callID, bool)` | `playbackControl.Mute()` / `.Unmute()` |
| Blind transfer | `Transfer(callID, target)` | `dialog.Refer(ctx, targetUri)` |
| Play WAV | `PlayAudio(callID, path)` | `dialog.PlaybackCreate(); pb.PlayFile(path)` |
| Record | `Record(callID, path)` | `dialog.AudioStereoRecordingCreate(file)` |

### SIP trace capture (MVP — validated in Phase 0 spike)

SIP trace capture is included in MVP. The Phase 0 spike validated sipgo's
`SIPTracer` interface as a first-class interception API (not a debug log
scraper). The approach:

1. Implement sipgo's `sip.SIPTracer` interface to intercept raw SIP messages
   with direction, transport, addresses, and timestamps.
2. Wrap intercepted messages as `SipTraceEvent` and push onto the events channel.
3. The TUI SIP Trace tab displays messages in real time (scrolling text view).
4. Feed events to a `DialogTracker` (Section 4) for the sngrep-style viewer
   in Phase 2.

**Hook point:** sipgo's `sip.SIPDebugTracer()` accepts a tracer implementation:

```go
sip.SIPDebug = true
sip.SIPDebugTracer(&engineTracer{events: engine.events})
```

Each `SIPTraceRead`/`SIPTraceWrite` callback receives the full raw SIP message
as `[]byte` plus transport metadata — sufficient for the trace log panel and
future dialog tracking.

---

## 4. Event System & State Management

The engine communicates with the TUI through a typed event system. Events are Go
structs sent over a single channel.

### Event types

```go
type RegStateEvent struct {
    AccountID string
    State     string  // "registered", "unregistered", "failed"
    Reason    string
}

type CallStateEvent struct {
    CallID    string
    State     string  // "calling", "incoming", "early", "confirmed", "disconnected"
    RemoteURI string
    Duration  time.Duration
    Direction string  // "inbound", "outbound"
}

type SipTraceEvent struct {
    Direction  string    // "send", "recv"
    Message    string    // full SIP message text
    Timestamp  time.Time
    CallID     string    // extracted for dialog grouping
    Method     string    // INVITE, REGISTER, BYE, etc.
    StatusCode int       // 0 for requests, 200/401/etc for responses
}

type DTMFEvent struct {
    CallID string
    Digit  rune
}
```

> **MVP note:** All four event types are wired end-to-end in MVP.
> `SipTraceEvent` is emitted by the engine's `SIPTracer` implementation and
> consumed by the TUI SIP Trace tab. `DTMFEvent` is wired for received DTMF
> display.

### TUI event loop

A single goroutine reads from `engine.Events()` and switches on event type:

```go
go func() {
    for ev := range engine.Events() {
        switch e := ev.(type) {
        case RegStateEvent:
            accountPanel.Update(e)
        case CallStateEvent:
            callTable.Update(e) // updates call table, not single-call display
        case DTMFEvent:
            callTable.ShowDTMF(e)
        case SipTraceEvent:
            traceLog.Append(e)
        }
        app.QueueUpdateDraw(func() {})
    }
}()
```

`app.QueueUpdateDraw()` is tview's thread-safe redraw trigger — similar in
purpose to Python Textual's `call_from_thread()` but simpler.

### Dialog tracker (deferred — not MVP)

The dialog tracker is a future component that ingests `SipTraceEvent`s and
groups them by Call-ID into dialog objects. This is the data model behind the
planned sngrep-style viewer. It is deferred until trace capture is validated
in the Phase 0 feasibility spike.

The planned data model is preserved here for reference:

```go
type DialogTracker struct {
    dialogs map[string]*Dialog  // keyed by Call-ID
    mu      sync.RWMutex
}

type Dialog struct {
    CallID   string
    From     string
    To       string
    State    string
    Messages []SipMessage  // ordered by timestamp
}
```

When implemented, SIP message parsing will extract enough from each trace event
to populate the dialog tracker — Call-ID, From, To, CSeq, method, status code.
sipgo has header parsing utilities that can be reused. The `internal/trace/`
package (Section 7) is reserved for this code.

---

## 5. TUI Layout & Navigation

The TUI uses tview's `Grid` for the three-column layout and `Pages` for the
tabbed bottom section.

### Main screen

```
┌──────────────────────────────────────────────────────────────────────┐
│ siptty                                           F1:Help  F10:Quit  │
├──────────────────┬───────────────────────────────┬───────────────────┤
│ ACCOUNTS         │ CALLS                         │ BLF / PRESENCE    │
│                  │                               │                   │
│ ● alice@pbx.io  │ ID   Remote         State Dur │ Ext  Name   State │
│   Registered     │ ▸ 1  sip:100@pbx   CONF 1:23 │ 201  Bob    ●Idle │
│   UDP 5060       │   2  sip:201@pbx   HOLD 0:45 │ 202  Carol  ◉Busy │
│                  │   3  sip:300@co     RING 0:03 │ 203  Dave   ○Off  │
│ ○ bob@sip.co    │                               │ 204  Eve    ◎Ring │
│   Unregistered   │ Dial: [___________________]   │                   │
│                  │                               │ [+] Add BLF...    │
│                  │ [d]Dial [a]Ans [h]Hang [m]Mute│                   │
├──────────────────┴───────────────────────────────┴───────────────────┤
│ [ Calls ] [ SIP Trace ] [ SIP Dialogs ]                             │
├──────────────────────────────────────────────────────────────────────┤
│ 15:04:01 INVITE sip:100@pbx.io → 100 Trying                        │
│ 15:04:01 INVITE sip:100@pbx.io → 180 Ringing                       │
│ 15:04:03 INVITE sip:100@pbx.io → 200 OK                            │
│ (SIP Dialogs tab is a placeholder — not active in MVP)               │
├──────────────────────────────────────────────────────────────────────┤
│ [d]Dial [a]Ans [h]Hang [x]Xfer [m]Mute [p]DTMF [Tab]Panels        │
└──────────────────────────────────────────────────────────────────────┘
```

### tview widget mapping

| UI Element | tview Widget | Notes |
|------------|-------------|-------|
| Overall layout | `Grid` (rows: header, main, tabs, footer) | |
| Three columns | `Grid` (3 columns: 1fr / 2fr / 1fr) | |
| Account list | `List` with colored items | |
| Call table | `Table` with selectable rows | Columns: ID, Remote, State, Duration. Supports multiple rows from day 1. MVP may limit engine to one active call, but UI shows all. |
| Dial input | `InputField` | Below call table |
| BLF panel | `Table` with colored rows | Columns: Extension, Name, State. Placeholder data in MVP — wired to SUBSCRIBE/NOTIFY in Phase 2. Colored state indicators: idle=green, busy=red, ringing=yellow, offline=grey. |
| Bottom tabs | `Pages` (switched by Tab key) | Three pages: Calls, SIP Trace, SIP Dialogs |
| Calls tab | `Table` — call event log | Summary of call events (not raw SIP) |
| SIP Trace tab | `TextView` (append-only, scrolling) | **Active in MVP** — real-time display of raw SIP messages via sipgo `SIPTracer` |
| SIP Dialogs tab | `TextView` | **Placeholder in MVP** — shows "SIP dialog viewer available in a future release" |
| Key hints | `TextView` anchored to bottom row | |

### SIP dialog viewer (sngrep-style) — deferred, not MVP

The sngrep-style dialog viewer is deferred until SIP trace capture is validated
and implemented. The design below is preserved as the target specification for
the future implementation. The "SIP Dialogs" tab in the TUI is a placeholder
that will be activated when this feature is built.

Implementation reference: https://github.com/irontec/sngrep

**Level 1 — Dialog list** (the "SIP Dialogs" tab): A `Table` showing all
captured SIP dialogs. Arrow up/down to navigate. Columns: Call-ID, From, To,
State, Message Count.

```
  Call-ID                     From          To        State       Msgs
▸ 98asd7f@10.0.0.5          alice@pbx     100@pbx   Confirmed      8
  bc3ef21@10.0.0.5          alice@pbx     201@pbx   Terminated     6
```

**Level 2 — Ladder diagram** (press Enter on a dialog): Replaces the tab
content with an ASCII ladder diagram. Arrow up/down highlights individual
messages. Esc returns to the dialog list.

```
  10.0.0.5:5060                        10.0.0.1:5060
  (alice)                              (pbx)
       |                                    |
       |──── INVITE sip:100@pbx ──────────▸|   ◂ selected
       |                                    |
       |◂─── 100 Trying ───────────────────|
       |◂─── 180 Ringing ──────────────────|
       |                                    |
       |◂─── 200 OK ───────────────────────|
       |──── ACK ──────────────────────────▸|
```

**Level 3 — Packet detail** (press Enter on a message): A modal overlay
showing the full SIP message with syntax highlighting. Esc closes it.

---

## 6. Configuration

Same TOML format as the Python design. An engineer's existing config file works
unchanged.

### Config structs

```go
type Config struct {
    General  GeneralConfig   `toml:"general"`
    Accounts []AccountConfig `toml:"accounts"`
    Audio    AudioConfig     `toml:"audio"`
}

type GeneralConfig struct {
    LogLevel  int    `toml:"log_level"`   // 0-6
    LogFile   string `toml:"log_file"`
    UserAgent string `toml:"user_agent"`  // default: "siptty/0.1"
}

type AccountConfig struct {
    Name         string `toml:"name"`
    Enabled      bool   `toml:"enabled"`       // default: true
    SipURI       string `toml:"sip_uri"`        // required
    AuthUser     string `toml:"auth_user"`      // defaults to user part of sip_uri
    AuthPassword string `toml:"auth_password"`
    Registrar    string `toml:"registrar"`      // required
    Transport    string `toml:"transport"`      // udp, tcp, tls
    Register     bool   `toml:"register"`       // default: true
    RegExpiry    int    `toml:"reg_expiry"`     // default: 300
    Headers      map[string]string `toml:"headers"` // custom header overrides
}

type AudioConfig struct {
    Mode      string `toml:"mode"`       // "null" or "file"
    PlayFile  string `toml:"play_file"`
    RecordDir string `toml:"record_dir"`
}
```

### Config file search order

1. CLI flag: `siptty --config /path/to/config.toml`
2. Current directory: `./siptty.toml`
3. Platform-native default via `os.UserConfigDir()`:

| Platform | Config location |
|----------|----------------|
| Linux | `~/.config/siptty/config.toml` |
| macOS | `~/Library/Application Support/siptty/config.toml` |
| Windows | `%APPDATA%\siptty\config.toml` |

### Minimal working config

```toml
[[accounts]]
name = "alice"
sip_uri = "sip:alice@pbx.example.com"
auth_password = "secret123"
registrar = "sip:pbx.example.com"
```

Everything else has sensible defaults: UDP transport, log level 3, null audio
mode, 300-second registration expiry, `auth_user` derived from the SIP URI.

Header overrides work at the config level — per-account `[accounts.headers]`
applies to all outgoing requests. The runtime header editor (F3) and per-request
overrides come in a future phase.

---

## 7. Project Structure & Build

### Directory layout

```
siptty/
├── _python-prototype/             # Archived Python+PJSIP prototype
│   ├── src/siptty/
│   ├── tests/
│   ├── docker/
│   ├── scripts/
│   ├── research/
│   ├── pyproject.toml
│   └── Makefile
├── cmd/
│   └── siptty/
│       └── main.go              # entry point, config loading, wires engine → TUI
├── spike/
│   └── main.go                  # Phase 0 feasibility spike (temporary, not shipped)
├── internal/
│   ├── config/
│   │   ├── config.go            # Config structs + TOML loader
│   │   └── config_test.go
│   ├── engine/
│   │   ├── engine.go            # Engine struct, Dial/Answer/Hangup/DTMF methods
│   │   ├── events.go            # Event type definitions
│   │   ├── account.go           # Account registration lifecycle
│   │   ├── call.go              # Call struct wrapping diago dialog sessions
│   │   └── engine_test.go
│   ├── trace/                   # DEFERRED — reserved for post-spike implementation
│   │   ├── parser.go            # SIP message header parser
│   │   ├── tracker.go           # DialogTracker — groups messages by Call-ID
│   │   └── tracker_test.go
│   └── tui/
│       ├── app.go               # tview App setup, layout, key bindings
│       ├── account_panel.go     # Account list widget
│       ├── call_panel.go        # Call control + dial input widget
│       ├── trace_panel.go       # SIP trace log (scrolling text)
│       ├── dialog_list.go       # Level 1: dialog table
│       ├── dialog_ladder.go     # Level 2: ladder diagram renderer
│       ├── dialog_detail.go     # Level 3: full message modal
│       └── app_test.go
├── go.mod
├── go.sum
├── Makefile
├── Docs/                        # design docs (shared)
├── tests/
│   ├── asterisk/                # Asterisk config (shared with Go tests)
│   └── integration_test.go      # integration tests against Asterisk in Docker
├── docker-compose.test.yml      # Asterisk test environment (shared)
└── README.md
```

### Key decisions

- `internal/` keeps all packages private. This is a tool, not a library.
- `cmd/siptty/main.go` is thin: parse flags, load config, create engine,
  create TUI, wire together, run.
- `trace/` is separate from `engine/` because the dialog tracker and SIP parser
  are pure data structures with no diago dependency — easy to unit test. This
  package is reserved but not implemented until the Phase 0 spike validates
  sipgo's message interception capabilities.
- `spike/` contains the Phase 0 feasibility spike — a standalone program that
  validates core assumptions before committing to the full build.
- Integration tests use `testcontainers-go` for Asterisk container lifecycle.

### Dependencies (go.mod)

```
github.com/emiago/sipgo      # SIP stack
github.com/emiago/diago      # call control + RTP
github.com/rivo/tview        # TUI framework
github.com/gdamore/tcell/v2  # terminal rendering (tview dependency)
github.com/BurntSushi/toml   # config parsing
```

Five dependencies. No CGO.

### Makefile targets

```makefile
build:            go build -o siptty ./cmd/siptty
test:             go test ./...
test-integration: docker compose -f docker-compose.test.yml up -d && \
                  go test -tags integration ./tests/
lint:             golangci-lint run
release:          # cross-compile for all platforms
```

---

## 8. Testing Strategy

### Unit tests

No network, no Docker, no diago. Run with `go test ./...`:

- **Config parsing:** Valid TOML loads, missing required fields error, defaults
  applied, invalid transport rejected, header overrides parsed.
- **SIP trace parser:** *(Deferred — not MVP)* Extract Call-ID, Method, status
  code, From, To from raw SIP message strings. Table-driven tests with real SIP
  message samples. Implemented when `internal/trace/` package is built.
- **Dialog tracker:** *(Deferred — not MVP)* Ingest a sequence of
  `SipTraceEvent`s, verify dialogs grouped correctly by Call-ID, state
  transitions tracked, message ordering preserved. Implemented when
  `internal/trace/` package is built.
- **Event types:** All event structs constructible with expected fields.
- **DTMF validation:** Valid digits accepted (0-9, *, #, A-D), invalid chars
  rejected.

### TUI tests

tview lacks a test pilot like Textual, so TUI testing is more limited:

- **Widget construction:** Panels create without panic, layout renders at a
  given terminal size.
- **Dialog ladder renderer:** *(Deferred — not MVP)* Given a `Dialog` struct,
  verify ASCII ladder output is correct. Pure function (data in, string out),
  easily testable. Implemented alongside the SIP dialog viewer.
- **Call table rendering:** Given a list of `CallStateEvent`s, verify the call
  table displays correct rows with expected columns (ID, Remote, State, Duration).
- **BLF panel rendering:** Given mock BLF state data, verify the panel displays
  correct rows with colored state indicators.
- **Key binding mapping:** Key constants map to expected action strings.

### Integration tests

Guarded by `//go:build integration` build tag. Require Docker. Reuse existing
Asterisk config files (`tests/asterisk/pjsip.conf`, `extensions.conf`,
`voicemail.conf`):

- **Registration:** Register against Asterisk, verify success event. Wrong
  password, verify failure.
- **Outbound call:** Dial echo extension (600), verify state progression
  through confirmed, hangup.
- **Inbound call:** Two engine instances (ext 100 + 101), one calls the other,
  answer, verify both confirmed, hangup.
- **DTMF:** Call DTMF test extension (602), send digits, verify no errors.
- **SIP trace capture:** *(Deferred — not MVP)* Register, verify trace events
  contain REGISTER and 200 OK. Implemented when trace capture is wired in.

Integration tests use `testcontainers-go` for container lifecycle. Set
`SIPTTY_ASTERISK_UP=1` to skip container management when Asterisk is already
running.

### Phase 0 feasibility spike tests (COMPLETED 2026-02-10)

The spike (`spike/main.go`) is a standalone program, not part of the test suite.
It validated each capability against a running Asterisk instance (Docker) and
all 6 tests passed. Full results: `Docs/Go-gosip-diago-summary.md`.

1. Registration against Asterisk via diago — **PASS** — `RegisterTransaction`
   with digest auth, register + unregister.
2. Outbound call — **PASS** — INVITE to ext 603 (Answer+Wait), verify 200 OK, BYE.
3. DTMF sending — **PASS** — call ext 603, send digits 1-4 via `AudioWriterDTMF`.
4. WAV file playback — **PASS** — programmatic 8kHz WAV, `PlaybackCreate().PlayFile()`.
5. Mute/unmute — **PASS** — `PlaybackControlCreate().Mute(true/false)` + `Stop()`.
6. Raw SIP message interception — **PASS** — sipgo `SIPTracer` interface provides
   full raw message capture. 66 messages captured across all tests, including
   REGISTER, INVITE, 100/200/401, ACK, BYE in both directions.

No "(risk)" items remain. All spike-validated features are confirmed for MVP.

---

## 9. MVP Scope & Deferred Work

### Phase 0 — Feasibility Spike (COMPLETED 2026-02-10)

The spike (`spike/main.go`) validated all six capabilities against Asterisk in
Docker using sipgo v1.2.0 and diago v0.27.0. Full results are documented in
`Docs/Go-gosip-diago-summary.md`.

| Validation target | What to prove | Result |
|---|---|---|
| Registration | `diago.Register()` against Asterisk, digest auth, refresh | **PASS** — `RegisterTransaction` with digest auth |
| Outbound call | `diago.Invite()` → SDP → RTP → `Hangup()` | **PASS** — full INVITE flow with auth challenge |
| DTMF sending | `dialog.AudioWriterDTMF().WriteDTMF()` | **PASS** — RFC 4733 telephone-events |
| WAV playback | `dialog.PlaybackCreate(); pb.PlayFile(path)` | **PASS** — 32KB to RTP from 2s WAV |
| Mute/unmute | `PlaybackControl.Mute()` / `.Unmute()` | **PASS** — silence frames, session stays active |
| Raw SIP interception | Access sipgo transport layer or logging hooks for raw messages | **PASS** — `sip.SIPTracer` interface provides full raw message capture with direction, transport, and addresses |

**Exit criteria: ALL MET.** No workarounds or fallbacks needed for any item.
The SIP trace interception is better than expected — `sip.SIPTracer` is a
first-class API, not a debug log scraper. SIP trace features can be promoted
from Phase 2 to Phase 1 if desired.

### Phase 1 — MVP

Everything diago provides out of the box, plus the multi-call TUI. SIP trace
capture and the dialog viewer are explicitly excluded from MVP.

| Feature | Source | Notes |
|---|---|---|
| SIP registration with digest auth | diago `Register()` | |
| Unregistration | diago context cancellation | |
| Outbound calls (INVITE/BYE) | diago `Invite()` | |
| Inbound calls (answer/reject) | diago `Serve()` + `Answer()` | |
| Hangup | diago `Hangup()` | |
| Send DTMF (RFC 4733) | diago `AudioWriterDTMF()` | Validated in spike |
| Receive DTMF | diago `AudioReaderDTMF()` | |
| SIP trace log panel | sipgo `SIPTracer` API | Validated in spike; real-time raw SIP display |
| Blind transfer (REFER) | diago `Refer()` | |
| File audio — play WAV into call | diago `PlaybackCreate()` | Validated in spike |
| File audio — record from call | diago `AudioStereoRecordingCreate()` | |
| Null audio mode (signaling only) | No media reader/writer | |
| TOML config file | `BurntSushi/toml` | |
| Custom SIP headers (config-level) | sipgo header access | |
| TUI with multi-call table | tview `Table` | UI supports N calls; engine may limit to 1 active |
| TUI with BLF placeholder panel | tview `Table` | Columns ready; no backend subscription yet |
| TUI with SIP Trace tab (active) | tview `Pages` | Real-time SIP message display; SIP Dialogs tab placeholder |
| Single-binary cross-platform builds | `go build`, CGO_ENABLED=0 | |
| G.711 ulaw/alaw codecs | diago built-in, pure Go | |

**Deferred from MVP to Phase 2:**
- Mute/unmute — diago `PlaybackControl.Mute()` validated in spike; not needed
  in MVP because null/canned audio mode doesn't benefit from mute control
- SIP dialog viewer (sngrep-style) — depends on trace capture (now in MVP);
  remains P2 scope due to UI complexity (ladder diagrams, dialog tracking)

### Phase 2 — Standard (Desk phone parity + SIP trace)

| Feature | Reason deferred | Custom work required |
|---|---|---|
| Mute/unmute | Not needed with null/canned audio in MVP | diago `PlaybackControl.Mute()` — spike validated |
| SIP dialog viewer | UI complexity (ladder diagrams) | `internal/trace/` package + TUI rendering; trace capture in MVP |
| Hold/resume | No diago API | re-INVITE with SDP direction |
| BLF subscriptions (RFC 4235) | No diago SUBSCRIBE/NOTIFY | SUBSCRIBE/NOTIFY transaction support |
| MWI (RFC 3842) | SUBSCRIBE/NOTIFY pattern | Same as BLF |
| Presence (RFC 3856) | SUBSCRIBE/NOTIFY pattern | SUBSCRIBE/PUBLISH handling |
| Attended transfer | Only blind transfer in diago | Replaces handling |
| Conference (3+ party) | diago bridge is 2-party only | Audio mixing or external bridge |
| Multiple simultaneous calls (engine) | Engine complexity | Call manager + hold/resume integration |
| Runtime header editor (F3) | Config-level headers cover MVP | Per-request header mutation pipeline |

### Phase 3 — Advanced

| Feature | Reason deferred | Custom work required |
|---|---|---|
| Opus codec | Requires CGO or pure Go encoder | Codec implementation + SDP negotiation |
| G.722 codec | Only minimal Go transpilation exists | Same |
| ICE/STUN/TURN | Not in diago | NAT traversal stack |
| SRTP media | Not in diago | Encrypted media pipeline |
| Session timers (RFC 4028) | Not in diago | Timer handling |
| 100rel/PRACK (RFC 3262) | Not in sipgo | Custom SIP extension |
| TLS signaling | Depends on sipgo TLS support | Transport configuration |
| DNS SRV/NAPTR (RFC 3263) | Validate sipgo support | May be built-in |
| SIP MESSAGE (IM) | Not in diago | MESSAGE support |
| Call history (SQLite) | Nice-to-have | App-layer storage |
| Browser audio (WebSocket) | Dropped — terminal-first tool, not planned | WebSocket audio bridge — not planned |

### Canonical Feature Mapping (DESIGN.md → Go Plan)

This table is the authoritative roadmap for feature parity with the original
Python/PJSIP design. It incorporates the reviewer's analysis of what sipgo/diago
provide and what requires custom implementation.

**Legend:**
- **MVP** = Phase 1 scope
- **P2** = Phase 2 scope
- **P3** = Phase 3 scope
- **Custom** = requires substantial custom SIP/media work beyond sipgo/diago
- **Dropped** = not planned for the Go rewrite
- **(risk)** = depends on spike validation or diago API limitations

| Feature (from DESIGN.md) | Go plan status | Notes |
|---|---|---|
| Register / Unregister | MVP | diago `RegisterTransaction()` — spike validated |
| Outbound / inbound calls | MVP | diago `Invite()` / `Serve()` — spike validated |
| Hangup | MVP | diago `Hangup()` — spike validated |
| Call hold / resume | P2 + Custom | Needs re-INVITE with SDP direction |
| Mute / unmute | P2 | diago `PlaybackControlCreate().Mute()` — spike validated; deferred (not needed with null/canned audio) |
| Send DTMF | MVP | diago `AudioWriterDTMF().WriteDTMF()` — spike validated |
| Receive DTMF | MVP | diago `AudioReaderDTMF()` — not spike-tested but same API surface |
| SIP trace display | MVP | sipgo `SIPTracer` interface validated in spike; promoted to MVP |
| SIP dialog viewer | P2 | Depends on trace capture — now validated, no risk |
| File audio play/record | MVP | diago `PlaybackCreate().PlayFile()` — spike validated |
| Null audio mode | MVP | No media reader/writer |
| BLF subscriptions | P2 + Custom | SUBSCRIBE/NOTIFY support needed |
| MWI | P2 + Custom | SUBSCRIBE/NOTIFY support needed |
| Presence pub/sub | P2 + Custom | SUBSCRIBE/PUBLISH handling |
| Blind transfer | MVP | diago `Refer()` |
| Attended transfer | P2 + Custom | Needs Replaces handling |
| 3-way conference | P2 + Custom | Audio mixing/bridge needed |
| Multiple simultaneous calls (UI) | MVP | TUI call table supports N rows from day 1 |
| Multiple simultaneous calls (engine) | P2 | Engine call manager + hold/resume |
| Call history (SQLite) | P3 | App-layer only |
| Runtime header editor | P2 + Custom | Per-request header mutation pipeline |
| Config header overrides | MVP | Config-level only |
| User-Agent spoofing | MVP | Header override or transport option |
| TLS signaling | P3 | Depends on sipgo TLS support |
| SRTP media | P3 + Custom | Not in diago |
| ICE/STUN/TURN | P3 + Custom | Not in diago |
| DNS SRV/NAPTR | P3 | sipgo RFC 3263 support — validate |
| 100rel/PRACK | P3 + Custom | Not in sipgo |
| SIP MESSAGE (IM) | P3 + Custom | Needs MESSAGE support |
| Browser audio | Dropped | Terminal-first tool; not a priority |

**Key changes from the reviewer's original table:**
- SIP trace display promoted to MVP (sipgo `SIPTracer` validated in spike)
- Mute/unmute demoted from MVP to P2 (not needed with null/canned audio)
- SIP dialog viewer remains P2 (UI complexity; trace capture now in MVP)
- Multiple simultaneous calls split into UI (MVP) and engine (P2)
- Runtime header editor moved from P3 to P2 (aligns with desk phone parity)
- All "(risk)" items de-risked by Phase 0 spike (completed 2026-02-10)
