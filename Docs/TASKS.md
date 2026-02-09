# siptty — Task List

## Phase 1: Design Documents
- [x] 1.1 Write DESIGN.md Part 1: Overview, goals, tech stack, architecture diagram
- [x] 1.2 Write DESIGN.md Part 2: Feature list (tiered: MVP → Standard → Advanced)
- [x] 1.3 Write DESIGN.md Part 3: TUI layout and screen designs
- [x] 1.4 Write DESIGN.md Part 4: pjsua2 integration architecture
- [x] 1.5 Write DESIGN.md Part 5: Configuration file format (TOML)
- [x] 1.6 Write DESIGN.md Part 6: BLF/presence subscription engine
- [x] 1.7 Write DESIGN.md Part 7: SIP header override system
- [x] 1.8 Write TEST_PLAN.md: Testing strategy with Asterisk Docker

---

## Phase 2: Project Setup

Three parallel tracks after the project skeleton (2.1) is done:

```
 Track A (project)          Track B (pjsua2)          Track C (docker)
 ──────────────────         ──────────────────        ──────────────────
 2.1.x src layout &
       pyproject.toml
         │
         ├────────────────▶ 2.2.x build/install        2.3.x compose +
         │                       pjsua2                      Asterisk config
         │
         ◀──────────── all three tracks merge ──────────────▶
         │
       2.4.x Textual app skeleton + smoke tests
```

### 2.1 Initialize Python project

- [ ] **2.1.1** Create src layout directory structure
  - `src/siptty/__init__.py` (with `__version__`), `__main__.py`,
    `tui/__init__.py`, `tui/app.py`, `tui/widgets/__init__.py`,
    `engine/__init__.py`, `config/__init__.py`
  - **Done:** `find src -name '*.py'` shows all expected files

- [ ] **2.1.2** Create `pyproject.toml` with metadata and dependencies
  - Hatchling build backend, `requires-python >= 3.11`
  - Runtime deps: `textual>=0.50`
  - Optional deps: `[test]` (pytest, pytest-asyncio, pytest-timeout, coverage),
    `[dev]` (ruff, mypy, textual-dev)
  - Entry point: `siptty = "siptty.tui.app:main"`
  - pjsua2 deliberately **not** listed (system-level build; see 2.2)
  - **Done:** `pip install -e '.[test,dev]'` succeeds; `siptty` is on PATH

- [ ] **2.1.3** Update `.gitignore` for Python build artifacts
  - Add: `.eggs/`, `src/*.egg-info/`, `.ruff_cache/`, `.mypy_cache/`, `.venv/`
  - Verify existing entries cover `__pycache__/`, `dist/`, `build/`
  - **Done:** `git status` clean after full install + lint cycle

- [ ] **2.1.4** Create entry points (`__main__.py` and `app.py:main()`)
  - `__main__.py` calls `from siptty.tui.app import main; main()`
  - `main()` prints placeholder and exits
  - **Done:** `python -m siptty` and `siptty` both run and exit 0

- [ ] **2.1.5** Add `py.typed` marker and linter/type-checker config
  - `src/siptty/py.typed` (empty marker)
  - `[tool.ruff]` in pyproject.toml (line-length=99, target py311)
  - `[tool.mypy]` in pyproject.toml (strict=true, py311)
  - **Done:** `ruff check src/` and `mypy src/` pass with zero errors

### 2.2 Get pjsua2 Python bindings building/installed

- [ ] **2.2.1** Document system-level build dependencies
  - Create `Docs/PJSUA2_INSTALL.md`
  - Required apt packages: `build-essential`, `python3-dev`, `swig`,
    `libasound2-dev`, `libopus-dev`, `libssl-dev`
  - Single copy-paste `apt install` one-liner for Ubuntu 24.04
  - **Done:** Doc exists with tested install instructions

- [ ] **2.2.2** Build and install pjsua2
  - Primary: `pip install pjsua2==2.12` (compiles from PyPI sdist)
  - Fallback: clone pjproject repo, `./configure && make`, install swig bindings
  - Document whichever path works; record build time
  - **Done:** `python -c "import pjsua2; print('OK')"` succeeds

- [ ] **2.2.3** Write verification script (`scripts/check_pjsua2.py`)
  - Creates Endpoint, calls `libCreate()`, prints `libVersion().full`,
    calls `libDestroy()`, exits 0
  - **Done:** `python scripts/check_pjsua2.py` prints version and exits 0

- [ ] **2.2.4** Add Makefile with install targets
  - Targets: `install-deps` (apt), `install-pjsua2` (pip), `check-pjsua2`
    (runs verify script), `install-dev` (`pip install -e '.[test,dev]'`)
  - **Done:** `make install-deps install-pjsua2 check-pjsua2` runs end-to-end

- [ ] **2.2.5** Make pjsua2 an optional import
  - `engine/__init__.py`: try/except ImportError sets `PJSUA2_AVAILABLE` flag
  - `conftest.py`: `requires_pjsua2` pytest marker auto-skips when absent
  - **Done:** `pytest tests/unit/` passes without pjsua2 installed;
    `pytest tests/integration/` skips cleanly

### 2.3 Asterisk Docker test environment

- [ ] **2.3.1** Create `docker-compose.test.yml`
  - Asterisk service (Alpine-based image), ports 5060/udp+tcp, 10000-10100/udp
  - Volume-mount `tests/asterisk/` → `/etc/asterisk/`
  - Health check: `asterisk -rx 'core show version'`
  - **Done:** `docker compose -f docker-compose.test.yml up -d` starts container

- [ ] **2.3.2** Write `tests/asterisk/pjsip.conf` with test extensions 100–110
  - UDP + TCP transports
  - 11 endpoints with auth (password = `test{ext}`), aors (max_contacts=5)
  - **Done:** `asterisk -rx 'pjsip show endpoints'` lists all 11

- [ ] **2.3.3** Write `tests/asterisk/extensions.conf` with test dialplan
  - Context `[testing]`: echo (600), MOH (601), DTMF (602), hold (603),
    transfer target (700), conference (800), ext-to-ext (`_1XX`), voicemail fallback
  - **Done:** `asterisk -rx 'dialplan show testing'` shows all extensions

- [ ] **2.3.4** Write `tests/asterisk/voicemail.conf`
  - `[default]` context, mailboxes 100–110
  - **Done:** `asterisk -rx 'voicemail show users'` lists 11 mailboxes

- [ ] **2.3.5** Write Asterisk readiness-check utility (`scripts/wait_for_asterisk.py`)
  - Sends SIP OPTIONS probe via raw UDP socket in retry loop
  - Accepts `--host`, `--port`, `--timeout` flags
  - **Done:** Exits 0 within seconds when Asterisk is running; exits 1 on timeout

- [ ] **2.3.6** Write `tests/conftest.py` with pytest fixtures
  - Session-scoped `asterisk` fixture: compose up, wait for ready, yield, compose down
  - `--asterisk-up` CLI flag to skip container lifecycle
  - Placeholder `engine` and `two_phones` fixtures (stubbed for Phase 3)
  - **Done:** `pytest tests/integration/ -v` manages Asterisk container lifecycle

- [ ] **2.3.7** Write one integration smoke test
  - `test_asterisk_responds_to_options`: raw UDP SIP OPTIONS → assert SIP/2.0 response
  - Validates Docker + Asterisk config + network without needing pjsua2
  - **Done:** `pytest tests/integration/test_smoke.py -v` passes green

### 2.4 Basic Textual app skeleton

- [ ] **2.4.1** Create `SipttyApp(App)` class
  - `TITLE`, `SUB_TITLE`, `CSS_PATH`, `BINDINGS` (quit: q/F10, help: F1)
  - `compose()` yields Header, three-column layout, TabbedContent, Footer
  - **Done:** `SipttyApp().run()` launches and renders without crash

- [ ] **2.4.2** Create Textual CSS file (`src/siptty/tui/siptty.tcss`)
  - Three-column layout (1fr / 2fr / 1fr), tabbed bottom section (1fr height)
  - Panel borders and titles
  - **Done:** Visually correct layout at 120×40 terminal size

- [ ] **2.4.3** Create placeholder widgets for each panel
  - `AccountPanel` — "ACCOUNTS" with placeholder text
  - `CallControlPanel` — "State: IDLE" + dial Input placeholder
  - `BlfPanel` — "BLF / PRESENCE" with placeholder text
  - **Done:** All three render headers and placeholder text; no import errors

- [ ] **2.4.4** Create tabbed bottom section with placeholder tabs
  - Tabs: "Calls", "SIP Trace" (RichLog), "Dialogs", "Config"
  - **Done:** Tab switching works via keyboard; each shows placeholder content

- [ ] **2.4.5** Wire entry point and verify end-to-end launch
  - `main()` calls `SipttyApp().run()`
  - Handle KeyboardInterrupt gracefully
  - **Done:** `siptty` and `python -m siptty` launch full TUI; q/F10 exits cleanly

- [ ] **2.4.6** Write Textual pilot smoke test
  - `test_app_launches_and_has_panels`: assert all panels and TabbedContent exist
  - `test_quit_key`: press q → app exits
  - **Done:** `pytest tests/tui/ -v` passes green

### 2.5 Docker build environment

- [x] **2.5.1** Create `Dockerfile` with multi-stage build
  - Stage 1 (builder): python:3.12-bookworm, install SWIG 4.1.x + build deps,
    compile pjsua2 from pjproject source
  - Stage 2 (runtime): python:3.12-slim-bookworm, copy compiled pjsua2,
    install siptty
  - **Done:** `docker build -t siptty .` succeeds; `docker run siptty --help` works

- [x] **2.5.2** Verify pjsua2 works inside the container
  - Run `scripts/check_pjsua2.py` inside the container
  - **Done:** `docker run siptty python scripts/check_pjsua2.py` prints version

- [ ] **2.5.3** Create `docker-compose.yml` for development
  - Mounts source code, uses the built image, enables hot-reload
  - **Done:** `docker compose up` starts siptty with live code

- [ ] **2.5.4** Update Makefile with Docker targets
  - `docker-build`: builds the image
  - `docker-run`: runs siptty in Docker with `--net=host`
  - `docker-test`: runs tests inside the container
  - **Done:** All three targets work

---

## Phase 3: MVP Implementation

Build order — each group proceeds once dependencies are met:

```
 Group A (config):      3.7 ─────────────────────────────────────────▶
 Group B (engine):      3.0 ───▶
 Group C (register):    3.1 ────────▶
 Group D (trace):       3.6 ─────▶         (parallel with E)
 Group E (outbound):    3.2 ────────────▶
 Group F (inbound):     3.3 ──────────▶
 Group G (hold):        3.4 ──────▶
 Group H (dtmf):        3.5 ──────▶
 Group I (file audio):  3.8 ──────▶        (after E, parallel with F)

 A ──▶ B ──▶ C ──▶ E ──▶ F ──▶ G ──▶ H
              │         │
              └──▶ D    └──▶ I
```

### 3.0 Engine Lifecycle & Foundation

Prerequisites for all other Phase 3 work. Not in the original task list but
essential — the engine needs to boot before it can do anything.

- [x] **3.0.1** Define event dataclasses in `siptty/engine/events.py`
  - `RegStateEvent` (account_id, state, reason)
  - `CallStateEvent` (call_id, state, remote_uri, duration, direction)
  - `SipTraceEvent` (direction, message, timestamp)
  - All frozen=True for thread safety
  - **Done:** All three importable with fields matching DESIGN.md

- [x] **3.0.2** Create `SipEngine` class skeleton with `start()` / `stop()`
  - `SipEngine(event_callback)` — callback receives events
  - `start(config)`: creates pjsua2 Endpoint, `libCreate()`, `libInit()`, `libStart()`
  - `stop()`: calls `libDestroy()`
  - **Done:** start → stop cycle completes without crash

- [x] **3.0.3** Create UDP transport in `start()`
  - `transportCreate(PJSIP_TRANSPORT_UDP, tp_cfg)` with ephemeral port
  - Store transport ID on engine
  - **Done:** Engine starts with a UDP transport

- [x] **3.0.4** Audio mode support (null and file)
  - Null mode: `Endpoint.audDevManager().setNullDev()` — signaling only
  - File mode: `AudioMediaPlayer` to play WAV into calls,
    `AudioMediaRecorder` to record received audio to WAV
  - Config field `audio.mode` selects mode ("null" or "file")
  - All tests use null audio mode
  - **Done:** Engine starts in null or file audio mode without errors

- [x] **3.0.5** Unit test: engine lifecycle
  - Start with null_audio=True, stop — no error
  - Double-stop is safe
  - Start with bad config raises cleanly
  - **Done:** `tests/unit/test_engine_lifecycle.py` passes

### 3.7 Config File Loading

Moved first in implementation order because the engine needs config to start.

- [x] **3.7.1** Define config dataclasses in `siptty/config/models.py`
  - `AppConfig`, `GeneralConfig`, `AccountConfig`
  - Fields match DESIGN.md §9 TOML schema
  - Sensible defaults for all optional fields
  - **Done:** Dataclasses importable; all fields documented

- [x] **3.7.2** Config loader: parse TOML into `AppConfig` using `tomllib`
  - `load_config(path) -> AppConfig`
  - Raises `ConfigError` with clear message on missing required fields
  - **Done:** Loads a valid TOML config and returns populated dataclass

- [x] **3.7.3** Config defaults and validation
  - Defaults: log_level=3, null_audio=False, transport="udp", reg_expiry=300
  - Validate transport ∈ {"udp", "tcp", "tls"}, sip_uri starts with `sip:`
  - auth_user defaults to user part of sip_uri if omitted
  - **Done:** All defaults applied; invalid values rejected with clear errors

- [x] **3.7.4** Unit tests for config loading
  - Load minimal (one account), load multi-account, defaults applied,
    invalid transport rejected, missing sip_uri raises ConfigError,
    empty file raises ConfigError, BLF + header overrides parsed
  - **Done:** `tests/unit/test_config.py` passes

### 3.1 SIP Registration (Single Account)

- [x] **3.1.1** Implement `PhoneAccount(pj.Account)` with `onRegState` callback
  - Maps `regIsActive`/`regStatus` → state string ("registered", "failed",
    "unregistered")
  - Constructs `RegStateEvent`, posts via `call_from_thread()`
  - **Done:** Callback fires and produces correct event dataclass

- [x] **3.1.2** Implement `SipEngine.add_account(cfg)`
  - Builds `pj.AccountConfig` from `AccountConfig` dataclass
  - Creates `PhoneAccount`, calls `account.create(acc_cfg)`
  - Stores in `_accounts` dict keyed by account name
  - **Done:** Account created and registration initiated

- [x] **3.1.3** Implement `SipEngine.remove_account(account_id)`
  - Calls `setRegistration(False)` then `shutdown()`
  - Removes from `_accounts`; no-op if not found
  - **Done:** Account unregisters and is cleaned up

- [x] **3.1.4** TUI account list widget showing registration state
  - `AccountPanel` displays each account name + state
  - Icons: ● green = registered, ○ red = failed, ◌ grey = unregistered
  - Updates reactively on `RegStateEvent`
  - **Done:** Widget reflects live registration state

- [x] **3.1.5** Wire startup: config → engine → accounts
  - `SipttyApp.on_mount()` loads config, calls `engine.start()`,
    loops accounts and calls `add_account()` for each enabled one
  - Config path from CLI arg or `~/.config/siptty/config.toml`
  - **Done:** App launches, engine starts, accounts register

- [ ] **3.1.6** Integration test: registration against Asterisk
  - Register success → `RegStateEvent(state="registered")`
  - Wrong password → `RegStateEvent(state="failed", reason contains "401")`
  - Unregister → `RegStateEvent(state="unregistered")`
  - **Done:** `tests/integration/test_registration.py` passes

### 3.2 Outbound Call (Originate, Hangup)

- [x] **3.2.1** Implement `PhoneCall(pj.Call)` with `onCallState` callback
  - Maps pjsua2 call states → `CallStateEvent` state strings
  - Includes remote_uri, direction="outbound", duration
  - Posts event via `call_from_thread()`
  - **Done:** Callback fires with correct state progression

- [x] **3.2.2** Implement `PhoneCall.onCallMediaState` — audio plumbing
  - On ACTIVE: connect call AudioMedia ↔ playback/capture devices
  - Works with null audio (null dev absorbs audio)
  - **Done:** Audio path established on call connect

- [x] **3.2.3** Implement `SipEngine.dial(account_id, uri, headers)`
  - Creates `PhoneCall`, builds `CallOpParam` (audioCount=1, videoCount=0)
  - Applies optional SIP headers via `SipTxOption`
  - Calls `makeCall(uri, prm)`, stores in `_calls` dict
  - **Done:** Outbound INVITE sent; call_id returned

- [x] **3.2.4** Implement `SipEngine.hangup(call_id)`
  - Looks up call, calls `call.hangup(prm)`
  - **Done:** BYE/CANCEL sent; call terminated

- [x] **3.2.5** TUI dial input and call state display
  - `d` focuses dial Input; Enter submits to `engine.dial()`
  - `CallStatusPanel` shows state, remote URI, duration
  - `h` calls `engine.hangup()` when call active
  - Status clears on DISCONNECTED
  - **Done:** Can dial and hang up from the TUI

- [ ] **3.2.6** Integration test: outbound call
  - Dial echo extension (600) → state progression → confirmed → hangup
  - Dial and cancel before answer → DISCONNECTED
  - Dial non-existent → DISCONNECTED with reason
  - **Done:** `tests/integration/test_outbound.py` passes

- [x] **3.2.7** Call cleanup on DISCONNECTED state
  - `onCallState` DISCONNECTED: remove from engine's `_calls` dict
  - No dangling references
  - **Done:** Calls are garbage-collected after disconnect

### 3.3 Inbound Call (Ring, Answer, Hangup)

- [x] **3.3.1** Implement `PhoneAccount.onIncomingCall` callback
  - Creates `PhoneCall` for incoming call ID
  - Posts `CallStateEvent(state="incoming", direction="inbound")`
  - **Done:** Incoming calls detected and event fired

- [x] **3.3.2** Implement `SipEngine.answer(call_id)`
  - Answers with 200 OK; media connected by `onCallMediaState`
  - **Done:** Incoming call answered; audio path established

- [x] **3.3.3** Implement `SipEngine.reject(call_id, code=486)`
  - Hangs up with given status code (486 Busy / 603 Decline)
  - **Done:** Incoming call rejected with proper SIP response

- [x] **3.3.4** TUI incoming call notification and answer/reject
  - On incoming: panel shows "INCOMING: {remote_uri}" with visual alert
  - `a` answers, `h` rejects
  - After answer: state transitions to CONFIRMED
  - **Done:** Can answer and reject calls from the TUI

- [ ] **3.3.5** Integration test: inbound call
  - Two engines (ext 100 + 101): B calls A, A answers, both confirmed, hangup
  - B calls A, A rejects → both DISCONNECTED
  - **Done:** `tests/integration/test_inbound.py` passes

### 3.4 Call Hold / Resume

- [x] **3.4.1** Implement `SipEngine.hold(call_id)`
  - Calls `call.setHold()`; fires `CallStateEvent(state="hold")`
  - **Done:** re-INVITE with sendonly sent

- [x] **3.4.2** Implement `SipEngine.resume(call_id)`
  - Calls `call.reinvite()` with UNHOLD flag
  - Fires `CallStateEvent(state="confirmed")`
  - **Done:** re-INVITE with sendrecv sent; audio reconnected

- [x] **3.4.3** TUI hold/resume toggle
  - Key binding toggles hold/resume
  - Panel shows "ON HOLD" when held; key hint updates
  - **Done:** Hold and resume work from TUI

- [ ] **3.4.4** Integration test: hold/resume
  - Call echo → confirmed → hold → verify hold event → resume → verify
    confirmed → hangup
  - **Done:** `tests/integration/test_hold.py` passes

### 3.5 DTMF Sending

- [x] **3.5.1** Implement `SipEngine.send_dtmf(call_id, digits)`
  - Validates digits (0-9, *, #, A-D only)
  - Calls `call.dialDtmf()` (RFC 4733 telephone-event)
  - **Done:** DTMF digits sent in-band

- [x] **3.5.2** TUI DTMF input mode
  - `p` enters DTMF mode; digit keys captured and sent
  - Visual feedback: "DTMF: 1234..." in call panel
  - Escape exits DTMF mode
  - **Done:** DTMF sendable from TUI during active call

- [x] **3.5.3** Unit test: DTMF validation
  - Valid digits accepted; invalid chars raise ValueError
  - **Done:** `tests/unit/test_dtmf.py` passes

- [ ] **3.5.4** Integration test: DTMF delivery
  - Call DTMF-test extension (602), send "1234"
  - Verify call stays CONFIRMED (no errors)
  - **Done:** `tests/integration/test_dtmf.py` passes

### 3.8 File Audio (Play and Record)

- [x] **3.8.1** Implement `SipEngine.play_audio(call_id, wav_path)`
  - Creates `AudioMediaPlayer`, opens WAV file
  - Connects player to call's `AudioMedia` via conference bridge
  - **Done:** WAV file audio sent into active call

- [x] **3.8.2** Implement `SipEngine.record_audio(call_id, wav_path)`
  - Creates `AudioMediaRecorder`, opens output WAV file
  - Connects call's `AudioMedia` to recorder via conference bridge
  - **Done:** Received audio saved to WAV file

- [x] **3.8.3** Implement `SipEngine.stop_audio(call_id)`
  - Disconnects and cleans up any active player/recorder on the call
  - Auto-cleanup on call disconnect
  - **Done:** Player/recorder stop cleanly

- [ ] **3.8.4** TUI file audio controls
  - When audio.mode is "file" and a call is active, show play/record controls
  - Key binding to start/stop playback and recording
  - Status indicator showing play/record state
  - **Done:** Can play WAV and record audio from TUI during a call

- [ ] **3.8.5** Integration test: file audio
  - Call echo extension, play WAV, record response, verify output file exists
    and has non-zero size
  - **Done:** `tests/integration/test_file_audio.py` passes

### 3.6 SIP Trace Log Panel

- [x] **3.6.1** Implement custom `pj.LogWriter` to capture SIP messages
  - Override `write(entry)`, filter for SIP messages
  - Construct `SipTraceEvent` with direction, message, timestamp
  - Post via `call_from_thread()`
  - **Done:** SIP messages captured as events

- [x] **3.6.2** Wire `LogWriter` into Endpoint config
  - Set `EpConfig.logConfig.writer`, level from config
  - Suppress console output, enable SIP message logging
  - **Done:** Engine captures SIP trace without stdout spam

- [x] **3.6.3** TUI SIP trace panel (RichLog widget)
  - "SIP Trace" tab in bottom panel
  - Format: timestamp + direction arrow + message
  - Auto-scroll; buffer limited to 1000 messages
  - **Done:** SIP messages visible in real time in the TUI

- [ ] **3.6.4** Integration test: trace capture
  - Register → verify SipTraceEvent with "REGISTER" (send)
    and "200 OK" (recv)
  - **Done:** `tests/integration/test_trace.py` passes

---

## Phase 4: Standard Features
- [ ] 4.1 Multiple accounts
- [ ] 4.2 Blind transfer
- [ ] 4.3 Attended transfer
- [ ] 4.4 3-way conference
- [ ] 4.5 BLF subscriptions
- [ ] 4.6 MWI (voicemail indicator)
- [ ] 4.7 Call history
- [ ] 4.8 Custom SIP header injection

## Phase 5: SIP Dialog Viewer (sngrep-style)
- [ ] 5.1 SIP message parser (extract method, status, Call-ID, From, To, CSeq)
- [ ] 5.2 Dialog tracker (group messages by Call-ID, track state)
- [ ] 5.3 Dialog list tab (DataTable: Call-ID, From, To, state, msg count)
- [ ] 5.4 Call flow ladder diagram (ASCII arrows between endpoints)
- [ ] 5.5 Message detail modal (full SIP message, syntax-highlighted)
- [ ] 5.6 Filter/search dialogs (by method, URI, Call-ID, state)
- [ ] 5.7 Export dialog (text/markdown dump)

## Phase 6: Advanced Features
- [ ] 6.1 TLS/SRTP
- [ ] 6.2 Presence publish/subscribe
- [ ] 6.3 Browser audio (WebSocket ↔ WebAudio bridge for live mic/speaker)
- [ ] 6.4 Codec selection/priority UI
- [ ] 6.5 NAT traversal (STUN/ICE/TURN)
- [ ] 6.6 DNS SRV/NAPTR
- [ ] 6.7 Auto-answer with header detection
- [ ] 6.8 PyInstaller packaging
