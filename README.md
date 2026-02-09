# Terminal Phone

A full-featured SIP softphone with a terminal UI. Built for telecom engineers,
devops, and anyone who needs a Swiss Army knife SIP tool over SSH or on a
headless box.

Think Bria or a Yealink desk phone, but in your terminal.

## Why

- **Test and debug SIP** — full SIP trace logging, override any header
  (To, From, User-Agent, custom X-headers) from config or at runtime
- **Emulate devices** — spoof User-Agent strings, manipulate Contact headers,
  emulate the SIP behavior of specific phones
- **BLF monitoring** — subscribe to extension state via dialog event package
  (RFC 4235), see who's on the phone in real time
- **Works over SSH** — no GUI, no browser, no X11 forwarding needed

## Features (Planned)

### MVP
- SIP registration with digest auth
- Outbound/inbound calls (answer, reject, hangup)
- Call hold/resume
- DTMF (RFC 4733)
- Mute/unmute
- Live SIP trace panel
- TOML config file
- Audio device selection

### Standard
- Multiple simultaneous accounts
- Blind and attended transfer
- 3-way conference (local mixing)
- BLF subscriptions (dialog event package)
- MWI (voicemail indicator)
- Call history (SQLite)
- Custom SIP header injection
- DND, auto-answer, call waiting

### Advanced
- TLS signaling + SRTP media encryption
- Presence publish/subscribe
- Call recording (WAV)
- Codec priority UI
- STUN/ICE/TURN NAT traversal
- Intercom/auto-answer with Alert-Info detection
- User-Agent spoofing
- Session timers, 100rel/PRACK
- SIP MESSAGE (instant messaging)
- PyInstaller single-binary packaging

## Tech Stack

| Layer | Choice |
|-------|--------|
| SIP + Media | [PJSIP](https://www.pjsip.org/) via pjsua2 Python bindings |
| Terminal UI | [Textual](https://github.com/Textualize/textual) |
| Config | TOML (Python 3.11+ stdlib) |
| Testing | pytest + Asterisk in Docker |
| Packaging | PyInstaller |

## TUI Preview

```
┌──────────────────────────────────────────────────────────────────────┐
│ Terminal Phone                                  F1:Help F10:Quit   │
├──────────────────┬───────────────────────────────┬──────────────────┤
│ ACCOUNTS         │ CALL CONTROL                  │ BLF / PRESENCE   │
│                  │                               │                  │
│ ● alice@pbx.io  │ State: IDLE                   │ ● 201 Available  │
│   Registered     │                               │ ◉ 202 On Phone   │
│   UDP 5060       │ Dial: [___________________]   │ ○ 203 Offline    │
│                  │                               │ ◎ 204 Ringing    │
│ ○ bob@sip.co    │ [d]Dial [a]Answer [h]Hangup   │ ● 205 Available  │
│   Unregistered   │ [x]Xfer [c]Conf   [m]Mute     │                  │
├──────────────────┴───────────────────────────────┴──────────────────┤
│ [ Calls ] [ SIP Trace ] [ History ] [ Config ]                     │
├────────────────────────────────────────────────────────────────────-┤
│ INVITE sip:100@pbx.io SIP/2.0                                      │
│ Via: SIP/2.0/UDP 10.0.0.5:5060;branch=z9hG4bK-524287-1             │
│ From: "Alice" <sip:alice@pbx.io>;tag=abc123                        │
│ To: <sip:100@pbx.io>                                               │
│ User-Agent: TerminalPhone/0.1                                      │
├────────────────────────────────────────────────────────────────────-┤
│ [d]Dial [a]Ans [h]Hang [x]Xfer [c]Conf [m]Mute [p]DTMF [F5]Trace  │
└────────────────────────────────────────────────────────────────────-┘
```

## Architecture

```
  Textual TUI (asyncio main thread)
        │
        │ call_from_thread()
        │
  SIP Engine Layer (Python wrapper)
        │
    pjsua2 / PJSIP (C, worker threads)
        │
   UDP/TCP/TLS + RTP
```

The TUI never touches pjsua2 directly. The engine layer translates between
pjsua2 callbacks (on PJSIP threads) and Textual's async event loop via
`call_from_thread()`. See [Docs/DESIGN.md](Docs/DESIGN.md) for full details.

## Project Structure

```
├── Docs/
│   ├── DESIGN.md       # Full application design document
│   ├── TEST_PLAN.md    # Testing strategy (Asterisk Docker, pytest)
│   └── TASKS.md        # Implementation task list
├── research/           # Feature research notes
│   ├── sip-softphone-features.md
│   ├── hardware-phone-features.md
│   ├── sip-rfcs.md
│   └── pjsip-textual-architecture.md
└── README.md
```

## Status

**Design phase complete.** Implementation not yet started.

See [Docs/TASKS.md](Docs/TASKS.md) for the roadmap.

## Prerequisites (for development)

- Python 3.11+
- PJSIP with Python bindings (`pjsua2`)
- Docker (for running Asterisk test environment)

## License

MIT
