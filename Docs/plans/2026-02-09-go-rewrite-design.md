# siptty Go Rewrite — Design Document

Date: 2026-02-09

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

### SIP trace capture

sipgo supports debug logging via environment variables. We hook into sipgo's
logging to intercept raw SIP messages, wrap them as `SipTraceEvent`, and push
onto the events channel.

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

### TUI event loop

A single goroutine reads from `engine.Events()` and switches on event type:

```go
go func() {
    for ev := range engine.Events() {
        switch e := ev.(type) {
        case RegStateEvent:
            accountPanel.Update(e)
        case CallStateEvent:
            callPanel.Update(e)
        case SipTraceEvent:
            traceLog.Append(e)
            dialogTracker.Ingest(e)
        }
        app.QueueUpdateDraw(func() {})
    }
}()
```

`app.QueueUpdateDraw()` is tview's thread-safe redraw trigger — similar in
purpose to Python Textual's `call_from_thread()` but simpler.

### Dialog tracker

A struct that ingests `SipTraceEvent`s and groups them by Call-ID into dialog
objects. This is the data model behind the sngrep-style viewer.

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

SIP message parsing extracts enough from each trace event to populate the dialog
tracker — Call-ID, From, To, CSeq, method, status code, Via, Contact.
Straightforward header parsing, not a full SIP parser. sipgo has header parsing
utilities we can reuse.

---

## 5. TUI Layout & Navigation

The TUI uses tview's `Grid` for the three-column layout and `Pages` for the
tabbed bottom section.

### Main screen

```
┌──────────────────────────────────────────────────────────────────────┐
│ siptty                                           F1:Help  F10:Quit  │
├──────────────────┬───────────────────────────────┬───────────────────┤
│ ACCOUNTS         │ CALL CONTROL                  │ BLF / PRESENCE    │
│                  │                               │                   │
│ ● alice@pbx.io  │ State: CONFIRMED              │ (placeholder for  │
│   Registered     │ Remote: sip:100@pbx.io        │  future phase)    │
│   UDP 5060       │ Duration: 00:01:23            │                   │
│                  │                               │                   │
│ ○ bob@sip.co    │ Dial: [___________________]   │                   │
│   Unregistered   │                               │                   │
│                  │ [d]Dial [a]Ans [h]Hang [m]Mute│                   │
├──────────────────┴───────────────────────────────┴───────────────────┤
│ [ Calls ] [ SIP Trace ] [ SIP Dialogs ]                             │
├──────────────────────────────────────────────────────────────────────┤
│ INVITE sip:100@pbx.io SIP/2.0                                      │
│ Via: SIP/2.0/UDP 10.0.0.5:5060;branch=z9hG4bK-524287-1            │
│ From: "Alice" <sip:alice@pbx.io>;tag=abc123                        │
│ ...                                                                  │
├──────────────────────────────────────────────────────────────────────┤
│ [d]Dial [a]Ans [h]Hang [x]Xfer [m]Mute [p]DTMF [Tab]Panels        │
└──────────────────────────────────────────────────────────────────────┘
```

### tview widget mapping

| UI Element | tview Widget |
|------------|-------------|
| Overall layout | `Grid` (rows: header, main, tabs, footer) |
| Three columns | `Grid` (3 columns: 1fr / 2fr / 1fr) |
| Account list | `List` with colored items |
| Call state | `TextView` with dynamic content |
| Dial input | `InputField` |
| BLF panel | `Table` (placeholder, wired in future phase) |
| Bottom tabs | `Pages` (switched by Tab key) |
| SIP trace log | `TextView` (append-only, scrolling) |
| Key hints | `TextView` anchored to bottom row |

### SIP dialog viewer (sngrep-style) — three levels of drill-down

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
├── cmd/
│   └── siptty/
│       └── main.go              # entry point, config loading, wires engine → TUI
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
│   ├── trace/
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
├── Docs/                        # design docs (carried over)
├── tests/
│   ├── asterisk/                # Asterisk config (carried over)
│   └── integration_test.go      # integration tests against Asterisk in Docker
└── docker-compose.test.yml      # Asterisk test environment (carried over)
```

### Key decisions

- `internal/` keeps all packages private. This is a tool, not a library.
- `cmd/siptty/main.go` is thin: parse flags, load config, create engine,
  create TUI, wire together, run.
- `trace/` is separate from `engine/` because the dialog tracker and SIP parser
  are pure data structures with no diago dependency — easy to unit test.
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
- **SIP trace parser:** Extract Call-ID, Method, status code, From, To from raw
  SIP message strings. Table-driven tests with real SIP message samples.
- **Dialog tracker:** Ingest a sequence of `SipTraceEvent`s, verify dialogs
  grouped correctly by Call-ID, state transitions tracked, message ordering
  preserved.
- **Event types:** All event structs constructible with expected fields.
- **DTMF validation:** Valid digits accepted (0-9, *, #, A-D), invalid chars
  rejected.

### TUI tests

tview lacks a test pilot like Textual, so TUI testing is more limited:

- **Widget construction:** Panels create without panic, layout renders at a
  given terminal size.
- **Dialog ladder renderer:** Given a `Dialog` struct, verify ASCII ladder
  output is correct. Pure function (data in, string out), easily testable.
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
- **SIP trace capture:** Register, verify trace events contain REGISTER and
  200 OK.

Integration tests use `testcontainers-go` for container lifecycle. Set
`SIPTTY_ASTERISK_UP=1` to skip container management when Asterisk is already
running.

---

## 9. MVP Scope & Deferred Work

### MVP — Phase 1

Everything diago provides out of the box, plus the TUI:

| Feature | Source |
|---------|--------|
| SIP registration with digest auth | diago `Register()` |
| Unregistration | diago context cancellation |
| Outbound calls (INVITE/BYE) | diago `Invite()` |
| Inbound calls (answer/reject) | diago `Serve()` + `Answer()` |
| Hangup | diago `Hangup()` |
| Send DTMF (RFC 4733) | diago `AudioWriterDTMF()` |
| Receive DTMF | diago `AudioReaderDTMF()` |
| Mute/unmute | diago `PlaybackControl.Mute()` |
| Blind transfer (REFER) | diago `Refer()` |
| File audio — play WAV into call | diago `PlaybackCreate()` |
| File audio — record from call | diago `AudioStereoRecordingCreate()` |
| Null audio mode (signaling only) | No media reader/writer |
| SIP trace log panel | sipgo debug logging hooks |
| SIP dialog viewer (sngrep-style) | Custom: trace parser + dialog tracker + tview |
| TOML config file | `BurntSushi/toml` |
| Custom SIP headers (config-level) | sipgo header access |
| TUI with full layout | tview Grid/Pages/Table |
| Single-binary cross-platform builds | `go build`, CGO_ENABLED=0 |
| G.711 ulaw/alaw codecs | diago built-in, pure Go |

### Phase 2 — Custom sipgo work required

| Feature | Reason deferred |
|---------|----------------|
| Hold/resume | No diago API — need re-INVITE with SDP direction |
| BLF subscriptions (RFC 4235) | No diago SUBSCRIBE/NOTIFY |
| MWI (RFC 3842) | SUBSCRIBE/NOTIFY pattern |
| Presence (RFC 3856) | SUBSCRIBE/NOTIFY pattern |
| Attended transfer | Only blind transfer in diago |
| Conference (3+ party) | diago bridge is 2-party only |
| Multiple simultaneous calls | UI + call manager complexity |

### Phase 3 — Advanced

| Feature | Reason deferred |
|---------|----------------|
| Opus codec | Requires CGO or pure Go encoder |
| G.722 codec | Only minimal Go transpilation exists |
| ICE/STUN/TURN | Not in diago |
| Session timers (RFC 4028) | Not in diago |
| 100rel/PRACK (RFC 3262) | Not in sipgo |
| Browser audio (WebSocket) | Significant custom work |
| Runtime header editor (F3) | Config-level headers cover MVP |
| Call history (SQLite) | Nice-to-have |
