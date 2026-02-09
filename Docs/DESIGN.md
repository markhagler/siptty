# Terminal Phone — Application Design

A full-featured SIP softphone with a terminal UI, built for telecom engineers,
devops, and anyone who needs a Swiss Army knife SIP tool on a headless box or
over SSH.

## 1. Goals

1. **Complete SIP UA** — register, call, transfer, conference, BLF, presence.
   Everything a Yealink desk phone or Bria can do, minus the touchscreen.
2. **Hackable SIP stack** — override any SIP header (To, From, Contact,
   User-Agent, custom X-headers) from config or at runtime. Ideal for testing,
   debugging, and emulating specific devices.
3. **Real-time TUI** — live registration state, call state, BLF panel, SIP
   trace log. Keyboard-driven, no mouse required.
4. **Single-binary distribution** — PyInstaller-packaged, download and run.
5. **Testable** — automated test suite against Asterisk in Docker.

## 2. Tech Stack

| Layer | Choice | Why |
|-------|--------|-----|
| SIP/Media | PJSIP via pjsua2 Python bindings | 20yr mature, full UA+RTP+codecs+ICE |
| TUI | Textual (Python) | Best-in-class terminal UI, CSS layout, async |
| Config | TOML | Human-friendly, good Python support |
| Packaging | PyInstaller | Proven single-binary for Python+native libs |
| Testing | pytest + Asterisk Docker | Real PBX integration tests |
| Language | Python 3.11+ | Sweet spot for pjsua2 bindings + Textual |

## 3. Architecture

```
┌────────────────────────────────────────────────────────┐
│                    Textual TUI (asyncio)                    │
│  ┌────────────┐ ┌─────────────┐ ┌─────────────┐     │
│  │ Accounts   │ │ Call Control │ │ BLF Panel   │     │
│  │ & Reg      │ │ & Dialpad   │ │ & Presence  │     │
│  └────────────┘ └─────────────┘ └─────────────┘     │
│  ┌──────────────────────────────────────────┐     │
│  │          SIP Trace / Debug Log            │     │
│  └──────────────────────────────────────────┘     │
└──────────────────────────┬─────────────────────────────┘
                          │
                 call_from_thread()
                          │
┌──────────────────────────┼─────────────────────────────┐
│                   SIP Engine Layer                          │
│  ┌────────────┐ ┌─────────────┐ ┌─────────────┐     │
│  │ Account    │ │ Call        │ │ Subscription│     │
│  │ Manager    │ │ Manager     │ │ Manager     │     │
│  └────────────┘ └─────────────┘ └─────────────┘     │
│  ┌────────────┐ ┌─────────────┐                     │
│  │ Header     │ │ Config      │                     │
│  │ Override   │ │ Loader      │                     │
│  └────────────┘ └─────────────┘                     │
└──────────────────────────┬─────────────────────────────┘
                          │
                    pjsua2 / PJSIP
                          │
                   UDP/TCP/TLS + RTP
```

### Key Architectural Decisions

- **Thread boundary**: PJSIP callbacks run on PJSIP worker threads.
  All UI updates cross into Textual's asyncio loop via `call_from_thread()`.
- **Engine layer**: Thin wrapper around pjsua2 that normalizes events into
  simple dataclasses the TUI consumes. The TUI never touches pjsua2 directly.
- **Subscription Manager**: Dedicated component for BLF, presence, and MWI
  subscriptions — not bolted onto the Account class.
- **Header Override**: A pipeline that intercepts outgoing SIP messages and
  applies configured overrides before they hit the wire.

## 4. Feature Tiers

### Tier 1 — MVP (Usable SIP phone)

| Feature | SIP Mechanism | pjsua2 API |
|---------|--------------|------------|
| Register to PBX | REGISTER + digest auth | Account, AccountConfig |
| Unregister | REGISTER w/ Expires: 0 | Account.setRegistration(False) |
| Make outbound call | INVITE | Call.makeCall() |
| Answer inbound call | 200 OK to INVITE | Call.answer() |
| Reject inbound call | 486/603 | Call.answer(CallOpParam) |
| Hang up | BYE / CANCEL | Call.hangup() |
| Call hold / resume | re-INVITE sendonly/recvonly | Call.setHold() / Call.reinvite() |
| Mute / unmute | Stop capture→call audio path | AudioMedia.stopTransmit() |
| Send DTMF | RFC 4733 telephone-event | Call.dialDtmf() |
| SIP trace display | Log callback | Endpoint.logWriter |
| Config file | TOML | tomllib (stdlib 3.11+) |
| Audio device select | PJSIP audio dev | Endpoint.audDevManager() |

### Tier 2 — Standard (Desk phone parity)

| Feature | SIP Mechanism | pjsua2 API |
|---------|--------------|------------|
| Multiple accounts | Multiple REGISTER | Multiple Account instances |
| Blind transfer | REFER | Call.xfer() |
| Attended transfer | REFER w/ Replaces | Call.xferReplaces() |
| 3-way conference | Local audio mixing | Conference bridge (AudioMedia) |
| BLF subscriptions | SUBSCRIBE dialog event (RFC 4235) | Buddy / custom SUBSCRIBE |
| MWI | SUBSCRIBE message-summary (RFC 3842) | Account.onMwiInfo() |
| Call history | Local storage | SQLite via Python |
| Custom header injection | Modify outgoing SIP msgs | SipTxOption + SipHeader |
| Do Not Disturb | Local call rejection | App logic |
| Auto-answer | Header detection (Alert-Info) | onIncomingCall() inspection |
| Call waiting | Multiple Call instances | UI + call management |
| Caller ID display | From header parsing | CallInfo.remoteUri |

### Tier 3 — Advanced (Power user / testing tool)

| Feature | SIP Mechanism | pjsua2 API |
|---------|--------------|------------|
| TLS signaling | SIPS transport | TransportConfig.tlsConfig |
| SRTP media | SDES key exchange (RFC 4568) | AccountConfig.mediaConfig |
| Presence pub/sub | SUBSCRIBE/PUBLISH presence (RFC 3856) | Buddy, PresConfig |
| Call recording | WAV capture | AudioMediaRecorder |
| Codec priority UI | SDP offer manipulation | Endpoint.codecSetPriority() |
| STUN/ICE/TURN | NAT traversal | AccountConfig.natConfig |
| DNS SRV/NAPTR | Server resolution (RFC 3263) | Built into PJSIP |
| Intercom/auto-answer | Alert-Info header trigger | onIncomingCall() header check |
| User-Agent spoofing | UA header override | EpConfig.uaConfig.userAgent |
| Arbitrary header edit | Runtime SIP header editor | SipTxOption per-request |
| Session timers | RFC 4028 | AccountConfig.callConfig |
| 100rel/PRACK | RFC 3262 | Built into PJSIP |
| SIP MESSAGE (IM) | RFC 3428 | Account.onInstantMessage() |
| Multicast page send | RTP to multicast addr | Direct socket / pjmedia |
| PyInstaller binary | Single-file packaging | Build pipeline |

## 5. TUI Layout

### Main Screen

```
┌──────────────────────────────────────────────────────────────────────┐
│ Terminal Phone                                  F1:Help F10:Quit │
├──────────────────┬───────────────────────────────┬──────────────────┤
│ ACCOUNTS         │ CALL CONTROL                  │ BLF / PRESENCE   │
│                  │                               │                  │
│ ● alice@pbx.io  │ State: IDLE                   │ ● 201 Available  │
│   Registered     │                               │ ◉ 202 On Phone   │
│   UDP 5060       │ Dial: [___________________]   │ ○ 203 Offline    │
│                  │                               │ ◎ 204 Ringing    │
│ ○ bob@sip.co    │ [d]Dial [a]Answer [h]Hangup  │ ● 205 Available  │
│   Unregistered   │ [x]Xfer [c]Conf   [m]Mute    │                  │
│                  │                               │ [+] Add BLF...   │
├──────────────────┴───────────────────────────────┴──────────────────┤
│ [ Calls ] [ SIP Trace ] [ History ] [ Config ]                      │
├──────────────────────────────────────────────────────────────────────┤
│ INVITE sip:100@pbx.io SIP/2.0                                      │
│ Via: SIP/2.0/UDP 10.0.0.5:5060;branch=z9hG4bK-524287-1           │
│ From: "Alice" <sip:alice@pbx.io>;tag=abc123                       │
│ To: <sip:100@pbx.io>                                              │
│ Call-ID: 98asd7f@10.0.0.5                                         │
│ CSeq: 1 INVITE                                                    │
│ Contact: <sip:alice@10.0.0.5:5060>                                │
│ Content-Type: application/sdp                                     │
│ User-Agent: TerminalPhone/0.1                                     │
│ ...                                                               │
├──────────────────────────────────────────────────────────────────────┤
│ [d]Dial [a]Ans [h]Hang [x]Xfer [c]Conf [m]Mute [p]DTMF [F5]Trace │
└──────────────────────────────────────────────────────────────────────┘
```

### Screens

| Screen | Trigger | Purpose |
|--------|---------|--------|
| Main | Default | Accounts + call control + BLF + log tabs |
| Dial | `d` | Focus dial input, type number, Enter to call |
| Transfer | `x` | Prompt for transfer target |
| DTMF Pad | `p` | Grid of 0-9,*,#,A-D keys for mid-call DTMF |
| Settings | `F2` | Edit config: accounts, codecs, headers, BLF |
| Add BLF | `+` on BLF panel | Enter extension to subscribe |
| Header Editor | `F3` | Edit SIP header overrides for next request |
| Call History | tab | DataTable of past calls |

### Textual Widget Mapping

| UI Element | Textual Widget |
|-----------|---------------|
| Account list | `ListView` with colored `ListItem`s |
| Call state display | `Static` with reactive updates |
| Dial input | `Input` |
| BLF panel | `DataTable` (ext, state, name) with colored rows |
| SIP trace log | `RichLog` (scrolling, syntax-highlighted) |
| Bottom tabs | `TabbedContent` |
| Key hints | `Footer` (auto-generated from BINDINGS) |
| Header editor | `ModalScreen` with `Input` fields |
| Settings | `Screen` with form widgets |

## 6. SIP Engine Layer

The engine is a Python module that wraps pjsua2, isolating all PJSIP
interaction from the TUI. The TUI communicates with the engine via:

- **Commands** (TUI → Engine): `register()`, `dial()`, `answer()`, `hangup()`,
  `hold()`, `transfer()`, `subscribe_blf()`, etc.
- **Events** (Engine → TUI): dataclass events posted via `call_from_thread()`.

### Event Types

```python
@dataclass
class RegStateEvent:
    account_id: str
    state: str          # "registered", "unregistered", "failed"
    reason: str

@dataclass
class CallStateEvent:
    call_id: int
    state: str          # "calling", "incoming", "early", "connecting",
                        #  "confirmed", "disconnected", "hold"
    remote_uri: str
    duration: float
    direction: str      # "inbound" | "outbound"

@dataclass
class BlfStateEvent:
    extension: str
    state: str          # "idle", "ringing", "busy", "offline"
    remote_uri: str | None

@dataclass 
class MwiEvent:
    account_id: str
    new_messages: int
    old_messages: int

@dataclass
class SipTraceEvent:
    direction: str      # "send" | "recv"
    message: str        # Full SIP message text
    timestamp: float
```

### Engine Classes

```python
class SipEngine:
    """Top-level engine. Owns Endpoint, accounts, subscriptions."""
    def start(config: AppConfig) -> None
    def stop() -> None
    def add_account(cfg: AccountConfig) -> str
    def remove_account(account_id: str) -> None
    def dial(account_id: str, uri: str, headers: dict) -> int
    def answer(call_id: int) -> None
    def hangup(call_id: int) -> None
    def hold(call_id: int) -> None
    def resume(call_id: int) -> None
    def blind_transfer(call_id: int, target: str) -> None
    def attended_transfer(call_id: int, other_call_id: int) -> None
    def send_dtmf(call_id: int, digits: str) -> None
    def mute(call_id: int, muted: bool) -> None
    def subscribe_blf(account_id: str, extension: str) -> None
    def unsubscribe_blf(extension: str) -> None
    def set_codec_priority(codec: str, priority: int) -> None

class PhoneAccount(pj.Account):
    """Wraps pjsua2 Account. Fires RegStateEvent, CallStateEvent."""

class PhoneCall(pj.Call):
    """Wraps pjsua2 Call. Fires CallStateEvent, media connect."""

class BlfSubscription(pj.Buddy):
    """Wraps pjsua2 Buddy for dialog-state SUBSCRIBE. Fires BlfStateEvent."""
```

## 7. BLF / Presence Subscription Engine

BLF is a first-class feature. The subscription manager:

1. **Config-driven**: BLF targets defined in TOML config per account.
2. **Runtime add/remove**: Add or remove BLF subscriptions from the TUI.
3. **SUBSCRIBE to dialog event package** (RFC 4235): Sends
   `SUBSCRIBE sip:ext@domain` with `Event: dialog`.
4. **Parse NOTIFY bodies**: XML `dialog-info` documents contain dialog state
   (trying, early, confirmed, terminated).
5. **Map to simple states**: idle, ringing, busy, offline.
6. **Display in BLF panel**: Color-coded rows in the TUI DataTable.
7. **Click-to-dial from BLF**: Press Enter on a BLF entry to call that ext.
8. **BLF pickup**: If ringing, option to do directed pickup (`*8ext` or
   server-specific feature code).

### pjsua2 BLF Implementation

```python
# pjsua2 Buddy class handles SUBSCRIBE/NOTIFY
class BlfSubscription(pj.Buddy):
    def __init__(self, account, extension, domain):
        super().__init__()
        cfg = pj.BuddyConfig()
        cfg.uri = f"sip:{extension}@{domain}"
        cfg.subscribe = True
        self.create(account, cfg)

    def onBuddyState(self):
        info = self.getInfo()
        # info.presMonitorEnabled, info.state, info.subState
        # Map to BlfStateEvent and post to TUI
```

Note: pjsua2's Buddy class uses the `presence` event package by default.
For true BLF (dialog event package, RFC 4235), we may need to use lower-level
pjsip APIs or configure the Buddy subscription with `Event: dialog`.
PJSIP supports this via `pjsua_buddy_config.subscribe` + custom event headers.

## 8. SIP Header Override System

The killer feature for testing/debugging. Three layers of header control:

### Layer 1: Config File (persistent defaults)
```toml
[headers]
User-Agent = "Yealink SIP-T54W 96.86.0.100"

[headers.invite]
X-Custom-Header = "value"
Alert-Info = "<http://example.com>;info=alert-autoanswer"

[headers.register]
X-Device-ID = "aa:bb:cc:dd:ee:ff"
```

### Layer 2: Runtime Header Editor (F3 screen)
A modal screen with key-value inputs. Edits apply to the next outgoing
request only, or can be pinned as persistent overrides.

### Layer 3: Per-Request (dial command)
When dialing, optional header overrides:
```
Dial: 100 --header "X-Foo: bar" --from "Spoofed <sip:fake@pbx.io>"
```

### Implementation
```python
class HeaderOverride:
    method_filter: str | None   # None=all, "INVITE", "REGISTER", etc.
    name: str                   # Header name
    value: str                  # Header value
    persistent: bool            # Survives across requests
    one_shot: bool              # Removed after first use

class HeaderPipeline:
    def apply(self, method: str, tx_option: pj.SipTxOption) -> None:
        for override in self.overrides:
            if override.method_filter is None or override.method_filter == method:
                hdr = pj.SipHeader()
                hdr.hName = override.name
                hdr.hValue = override.value
                tx_option.headers.append(hdr)
```

## 9. Configuration File Format

```toml
[general]
log_level = 3                    # 0-6, PJSIP log verbosity
log_file = "terminal-phone.log"
null_audio = false               # true = no sound device (signaling only mode)
user_agent = "TerminalPhone/0.1"

[[accounts]]
name = "alice"
enabled = true
sip_uri = "sip:alice@pbx.example.com"
auth_user = "alice"
auth_password = "secret123"
registrar = "sip:pbx.example.com"
outbound_proxy = ""              # optional
transport = "udp"                # udp | tcp | tls
register = true
reg_expiry = 300

  [accounts.nat]
  stun_server = "stun:stun.l.google.com:19302"
  ice_enabled = false
  turn_enabled = false
  turn_server = ""
  turn_user = ""
  turn_pass = ""

  [accounts.tls]
  cert_file = ""
  key_file = ""
  ca_file = ""
  verify_server = true

  [accounts.srtp]
  enabled = false
  require = false                # false=optional, true=mandatory

  [accounts.codecs]
  priority = ["opus/48000", "g722/16000", "pcmu/8000", "pcma/8000"]

  [[accounts.blf]]
  extension = "201"
  label = "Bob"

  [[accounts.blf]]
  extension = "202"
  label = "Carol"

  [accounts.headers]
  User-Agent = "Yealink SIP-T54W 96.86.0.100"
  X-Custom = "test"

[[accounts]]
name = "bob"
sip_uri = "sip:bob@other-pbx.com"
# ... second account ...

[audio]
input_device = ""                # empty = system default
output_device = ""
ring_file = "ring.wav"           # optional custom ring

[history]
enabled = true
db_file = "~/.terminal-phone/history.db"
max_entries = 1000
```
