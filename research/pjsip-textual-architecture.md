# PJSIP pjsua2 + Textual Architecture Notes

## pjsua2 Python API

### Class Hierarchy

```
pj.Endpoint          — Singleton, manages the SIP stack lifecycle
pj.AccountConfig     — Configuration for a SIP account
  .idUri             — AOR (sip:user@domain)
  .regConfig         — RegistrarURI, timeoutSec, retryIntervalSec
  .sipConfig         — AuthCreds, proxies, transport
  .mediaConfig       — SRTP settings
  .natConfig         — STUN, ICE, TURN settings
  .presConfig        — Presence/PUBLISH settings
pj.Account           — Subclass this; override onRegState(), onIncomingCall()
pj.CallOpParam       — Parameters for call operations
pj.CallSetting       — Audio/video count, flags
pj.Call              — Subclass this; override onCallState(), onCallMediaState(), onCallTransferRequest()
pj.CallInfo          — State, duration, remoteUri, etc. via call.getInfo()
pj.AudioMedia        — Audio port abstraction
pj.AudioMediaPlayer  — Play WAV files
pj.AudioMediaRecorder— Record to WAV
pj.TransportConfig   — Bind address, port, TLS settings
pj.SipHeader         — name + value pair
pj.SipTxOption       — Custom headers for outgoing requests
pj.ToneGenerator     — DTMF tone generation
pj.CodecInfo         — Codec priority configuration
```

### Event/Callback Model

```python
class MyAccount(pj.Account):
    def onRegState(self, prm):         # Registration state changed
    def onIncomingCall(self, prm):      # New incoming INVITE
    def onIncomingSubscribe(self, prm): # Incoming SUBSCRIBE
    def onMwiInfo(self, prm):           # MWI notification
    def onInstantMessage(self, prm):    # SIP MESSAGE received

class MyCall(pj.Call):
    def onCallState(self, prm):         # Call state changed (EARLY, CONFIRMED, DISCONNECTED)
    def onCallMediaState(self, prm):    # Media state changed (ACTIVE, HOLD)
    def onCallTransferRequest(self, prm): # REFER received
    def onCallReplaced(self, prm):      # Replaces header
    def onDtmfDigit(self, prm):         # DTMF digit received
    def onCallTsxState(self, prm):      # Low-level transaction state
    def onCallSdpCreated(self, prm):    # SDP created - can modify
```

### Header Manipulation

```python
# Outgoing custom headers
tx_opt = pj.SipTxOption()
hdr = pj.SipHeader()
hdr.hName = "X-Custom-Header"
hdr.hValue = "custom-value"
tx_opt.headers.append(hdr)

# User-Agent is set via Endpoint config
ep_cfg = pj.EpConfig()
ep_cfg.uaConfig.userAgent = "TerminalPhone/1.0"

# Modify From display name
acc_cfg = pj.AccountConfig()
acc_cfg.idUri = '"Display Name" <sip:user@domain>'
```

### Transport Configuration

```python
# UDP
tp_cfg = pj.TransportConfig()
tp_cfg.port = 5060
ep.transportCreate(pj.PJSIP_TRANSPORT_UDP, tp_cfg)

# TCP
tp_cfg.port = 5060
ep.transportCreate(pj.PJSIP_TRANSPORT_TCP, tp_cfg)

# TLS
tp_cfg.port = 5061
tp_cfg.tlsConfig.certFile = "cert.pem"
tp_cfg.tlsConfig.privKeyFile = "key.pem"
tp_cfg.tlsConfig.caListFile = "ca.pem"
ep.transportCreate(pj.PJSIP_TRANSPORT_TLS, tp_cfg)
```

### NAT/ICE Configuration

```python
acc_cfg.natConfig.iceEnabled = True
acc_cfg.natConfig.turnEnabled = True
acc_cfg.natConfig.turnServer = "turn:turn.example.com"
acc_cfg.natConfig.turnUserName = "user"
acc_cfg.natConfig.turnPassword = "pass"
acc_cfg.natConfig.sipStunUse = pj.PJSUA_STUN_USE_DEFAULT
acc_cfg.natConfig.mediaStunUse = pj.PJSUA_STUN_USE_DEFAULT

ep_cfg.uaConfig.stunServer.append("stun:stun.example.com")
```

### Media Operations

```python
# Connect call audio to sound device
def onCallMediaState(self, prm):
    ci = self.getInfo()
    for mi in ci.media:
        if mi.type == pj.PJMEDIA_TYPE_AUDIO and mi.status == pj.PJSUA_CALL_MEDIA_ACTIVE:
            m = self.getAudioMedia(mi.index)
            # Connect call to speaker
            m.startTransmit(pj.Endpoint.instance().audDevManager().getPlaybackDevMedia())
            # Connect microphone to call
            pj.Endpoint.instance().audDevManager().getCaptureDevMedia().startTransmit(m)

# Mute: stop capture → call transmit
# Hold: call.setHold()
# Conference: connect multiple AudioMedia together via conference bridge
```

### Threading Model

PJSIP runs its own worker threads. Callbacks fire on PJSIP threads.
Textual runs an asyncio event loop on the main thread.

**Integration pattern:**
```python
# In PJSIP callback (runs on PJSIP thread):
def onCallState(self, prm):
    info = self.getInfo()
    # Post to Textual's event loop
    self.app.call_from_thread(self.app.handle_call_state, info)

# In Textual app:
def handle_call_state(self, info):
    # Safe to update widgets here
    self.query_one("#call-status").update(str(info.state))
```

`call_from_thread()` is Textual's thread-safe way to schedule work on the main async loop.

## Textual TUI Framework

### Key Widgets
- `Header` / `Footer` — app chrome with title and key bindings
- `Static` / `Label` — text display
- `Input` — text input field
- `Button` — clickable button
- `DataTable` — sortable/scrollable table
- `RichLog` — scrolling log output (perfect for SIP trace)
- `Tree` — collapsible tree view
- `TabbedContent` / `TabPane` — tabbed interface
- `ListView` / `ListItem` — scrollable list
- `Switch` / `Checkbox` / `RadioSet` — toggles
- `Select` / `SelectionList` — dropdown/multi-select
- `ProgressBar` — progress indicator
- `Sparkline` — inline chart
- `Placeholder` — layout debugging
- `Markdown` / `MarkdownViewer` — render markdown
- `ContentSwitcher` — show/hide content panels
- `Collapsible` — expandable sections

### Layout System (CSS)
```css
/* Textual uses a subset of CSS for layout */
#main {
    layout: horizontal;  /* or vertical, grid */
    height: 1fr;         /* fractional units */
    width: 50%;          /* or auto, specific values */
    padding: 1 2;        /* cells */
    margin: 1;
    border: solid green;
    background: $surface;
}
```

### Screen System
- `App` has a default screen
- `Screen` subclasses for modals, settings, etc.
- `app.push_screen()` / `app.pop_screen()` for navigation
- `ModalScreen` for dialogs that overlay

### Reactivity
```python
class MyApp(App):
    call_state = reactive("idle")  # Auto-triggers watch method
    
    def watch_call_state(self, value):
        self.query_one("#status").update(value)
```

### Workers (Background Tasks)
```python
@work(thread=True)
def do_registration(self):
    # Runs in thread pool, can call self.app.call_from_thread()
    pass
```

### Key Bindings
```python
class MyApp(App):
    BINDINGS = [
        Binding("q", "quit", "Quit"),
        Binding("d", "dial", "Dial"),
        Binding("a", "answer", "Answer"),
        Binding("h", "hangup", "Hangup"),
        Binding("m", "mute", "Mute"),
        Binding("t", "transfer", "Transfer"),
    ]
```
