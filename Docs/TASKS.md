# Terminal Phone - Task List

## Phase 1: Design Documents
- [x] 1.1 Write DESIGN.md Part 1: Overview, goals, tech stack, architecture diagram
- [x] 1.2 Write DESIGN.md Part 2: Feature list (tiered: MVP → Standard → Advanced)
- [x] 1.3 Write DESIGN.md Part 3: TUI layout and screen designs
- [x] 1.4 Write DESIGN.md Part 4: pjsua2 integration architecture
- [x] 1.5 Write DESIGN.md Part 5: Configuration file format (TOML)
- [x] 1.6 Write DESIGN.md Part 6: BLF/presence subscription engine
- [x] 1.7 Write DESIGN.md Part 7: SIP header override system
- [x] 1.8 Write TEST_PLAN.md: Testing strategy with Asterisk Docker

## Phase 2: Project Setup
- [ ] 2.1 Initialize Python project (pyproject.toml, src layout)
- [ ] 2.2 Get pjsua2 Python bindings building/installed
- [ ] 2.3 Asterisk Docker test environment
- [ ] 2.4 Basic Textual app skeleton that launches

## Phase 3: MVP Implementation
- [ ] 3.1 SIP registration (single account)
- [ ] 3.2 Outbound call (originate, hangup)
- [ ] 3.3 Inbound call (ring, answer, hangup)
- [ ] 3.4 Call hold/resume
- [ ] 3.5 DTMF sending
- [ ] 3.6 SIP trace log panel
- [ ] 3.7 Config file loading

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
- [ ] 6.3 Call recording
- [ ] 6.4 Codec selection/priority UI
- [ ] 6.5 NAT traversal (STUN/ICE/TURN)
- [ ] 6.6 DNS SRV/NAPTR
- [ ] 6.7 Auto-answer with header detection
- [ ] 6.8 PyInstaller packaging
