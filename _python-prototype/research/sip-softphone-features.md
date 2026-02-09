# SIP Softphone Feature Research

Comprehensive feature analysis of major SIP softphone applications.
Compiled for siptty project reference.

---

## 1. Bria (CounterPath, now Alianza)

**Platform:** Windows, macOS, iOS, Android
**License:** Commercial (subscription)

### SIP Registration & Authentication
- Multiple simultaneous SIP account registrations
- SIP digest authentication (MD5, SHA-256)
- Registration expiry / refresh interval configuration
- Re-registration interval tuning
- Outbound proxy support
- SIP domain / realm configuration
- Auth username separate from SIP URI username
- Display name configuration
- Transport selection per account (UDP / TCP / TLS)
- SIP keep-alive (OPTIONS ping, CRLF keep-alive, periodic REGISTER)
- Voicemail URI per account
- Registrar failover / redundancy (multiple proxies)
- SIP Outbound (RFC 5626) flow-token support

### Call Features
- Basic call lifecycle (INVITE, 180/183, 200 OK, ACK, BYE, CANCEL)
- Call hold / resume (re-INVITE with sendonly/recvonly SDP direction)
- Local music-on-hold playback
- Attended transfer (REFER with Replaces header)
- Blind / unattended transfer (REFER)
- 3-way local conference (local audio mixing)
- N-way conference (via external conference bridge URI)
- Call waiting / multiple concurrent call appearances
- Call forwarding (CFB, CFNA, CFU - server-side features)
- Do Not Disturb (local + server-side signaling)
- Auto-answer (configurable delay; header-triggered via Alert-Info / Call-Info)
- Call recording to local file (WAV)
- Call timer display
- Mute / unmute microphone
- Speaker volume control
- Microphone gain control
- Speed dial buttons
- Redial last number
- Click-to-call URL handlers (sip:, tel:, callto:)
- Drag-and-drop transfer (UI)
- Call park / retrieve (server-dependent features)
- Directed call pickup / group call pickup
- Intercom / paging (auto-answer triggered by Alert-Info header)
- Early media playback (183 Session Progress with SDP)
- 100rel / PRACK support (RFC 3262)
- Session timers (RFC 4028)
- UPDATE method (RFC 3311)

### Audio Codec Support
- G.711 u-law (PCMU) - 64 kbps
- G.711 A-law (PCMA) - 64 kbps
- G.729a / G.729ab - 8 kbps (licensed)
- G.722 - HD Voice wideband (48/56/64 kbps)
- Opus (narrowband to fullband, 6-510 kbps, variable bitrate)
- iLBC (13.33 / 15.2 kbps)
- GSM-FR (13 kbps)
- Speex (NB 8kHz, WB 16kHz, UWB 32kHz)
- Codec priority ordering / preference list
- SDP offer/answer codec negotiation
- VAD / silence suppression
- Comfort noise generation (CN, RFC 3389)

### Video Features
- H.264 (AVC)
- H.263 / H.263+
- VP8
- Video call with self-view preview
- Video resolution / bandwidth selection
- Screen sharing

### Network / Transport
- UDP transport
- TCP transport
- TLS transport (certificate validation, SNI)
- SIP over WebSocket (WSS) in newer versions
- IPv4 and IPv6 dual-stack
- DNS SRV record lookup
- DNS NAPTR record lookup
- A / AAAA record fallback
- Multiple network interface handling
- Proxy failover / redundancy
- Connection reuse (RFC 5765)

### NAT Traversal
- STUN client (RFC 5389)
- TURN client (RFC 5766)
- ICE full implementation (RFC 5245 / RFC 8445)
- ICE-Lite interoperability
- Symmetric RTP
- rport support (RFC 3581)
- Outbound proxy for NAT traversal
- UPnP IGD (optional)
- Configurable external / public IP address override

### Security
- TLS for SIP signaling (TLSv1.2, TLSv1.3)
- SRTP media encryption (RFC 3711) with SDES key exchange (RFC 4568)
- ZRTP media encryption (RFC 6189) with SAS verification
- Client certificate authentication (mutual TLS)
- CA certificate validation / custom CA bundles
- Certificate hostname verification
- Encrypted credential storage

### Presence & BLF
- SIP SUBSCRIBE/NOTIFY for presence (RFC 3856)
- Busy Lamp Field (BLF) via dialog event package (RFC 4235)
- Presence states: online, busy, away, on-the-phone, offline
- Custom presence status messages
- Watched / watcher user lists
- Combined speed-dial + BLF panel buttons
- XMPP presence integration

### Call History / Logs
- Full call history: incoming, outgoing, missed, declined
- Call duration per entry
- Call history search / filter
- Missed call notification (badge, popup)
- Call history export

### Contact Management
- Local contact storage
- LDAP directory integration
- XMPP roster integration
- Contact search / filter
- Contact groups / categories
- vCard import / export
- Corporate directory lookup
- Favorites list
- Contact photos
- OS native contacts integration (macOS Contacts, Outlook)

### DTMF
- RFC 2833 / RFC 4733 (telephone-event in RTP)
- SIP INFO method (application/dtmf-relay)
- Inband DTMF (audio tones in media stream)
- Configurable DTMF method per account
- On-screen dial pad for mid-call DTMF

### Voicemail
- MWI - Message Waiting Indicator (RFC 3842, SUBSCRIBE/NOTIFY)
- Visual indicator (icon / badge) for new voicemail
- One-touch voicemail dial button
- Voicemail number configuration per account

### Multiple Accounts / Lines
- Multiple simultaneous SIP registrations
- Per-account codec preferences
- Per-account transport selection
- Per-account proxy settings
- Default / preferred account selection
- Account enable / disable without deletion
- Multiple call appearances per account

### Provisioning
- Centralized provisioning via Stretto platform
- HTTPS-based auto-provisioning
- Configuration file import
- QR code provisioning (mobile)
- MAC-based provisioning
- Zero-touch deployment (enterprise)

### QoS / DSCP
- DSCP / ToS marking for SIP signaling packets
- DSCP / ToS marking for RTP media packets
- Separate DSCP values for audio vs video
- Adaptive jitter buffer
- Packet loss concealment (PLC)

### Diagnostics
- Full SIP message trace logging
- Configurable log levels (debug, info, warning, error)
- Log file export / save
- Audio device testing / diagnostics
- Real-time call quality indicators
- Call statistics: MOS, jitter, packet loss, RTT
- RTCP statistics display
- Network quality alerts

### Messaging
- SIP SIMPLE instant messaging (MESSAGE method)
- XMPP messaging
- Group messaging
- File transfer
- Desktop notifications
- SMS via SIP MESSAGE

### Advanced Features
- Custom ringtone per contact / per account
- Headset integration (Jabra, Plantronics/Poly, Sennheiser/EPOS) via vendor SDK
- Noise suppression / noise cancellation
- Acoustic echo cancellation (AEC)
- Automatic gain control (AGC)
- Push notifications (iOS APNs, Android FCM)
- Background operation (mobile)
- Hot-desking support
- CRM integration API
- Click-to-call browser extension

---

## 2. Linphone

**Platform:** Windows, macOS, Linux, iOS, Android, Web (WASM)
**License:** Open Source (GPLv3); commercial SDK license available
**SIP Stack:** liblinphone (belle-sip for SIP, mediastreamer2 for media)

### SIP Registration & Authentication
- Multiple simultaneous SIP account registrations
- SIP digest authentication (MD5, SHA-256)
- Registration expiry / refresh interval
- Outbound proxy support
- Auth username separate from SIP username
- Display name
- Transport per account (UDP / TCP / TLS / DTLS)
- SIP keep-alive (CRLF, OPTIONS)
- Contact header rewriting for NAT
- SIP Instance ID (RFC 5626 Outbound)
- GRUU support (RFC 5627)
- Registration backoff / retry logic
- Push notification gateway registration (mobile)

### Call Features
- Basic calling (INVITE / ACK / BYE / CANCEL)
- Call hold / resume (re-INVITE with sendonly/recvonly/inactive)
- Attended transfer (REFER with Replaces)
- Blind transfer (REFER)
- Local N-way conference calling (mediastreamer2 audio mixing)
- Server-based conference (focus URI)
- Call waiting / multiple simultaneous calls
- Do Not Disturb
- Auto-answer
- Call recording (WAV for audio, MKV for audio+video)
- Mute / unmute
- Volume control
- Call timer
- Early media (183 Session Progress with SDP)
- 100rel / PRACK (RFC 3262)
- Session timers (RFC 4028)
- UPDATE method (RFC 3311)
- Re-INVITE for codec / media renegotiation
- Call decline with reason (603 Decline, 486 Busy Here)
- 302 redirect / call forwarding
- Generic SIP header access via API

### Audio Codec Support
- G.711 u-law (PCMU)
- G.711 A-law (PCMA)
- G.722 (wideband)
- Opus (narrowband to fullband, variable bitrate)
- Speex (NB 8kHz, WB 16kHz, UWB 32kHz)
- iLBC
- GSM-FR
- SILK
- AAC-ELD
- G.729 (patent-free implementation / plugin)
- BV16 (BroadVoice16)
- L16 (linear PCM)
- Codec priority ordering
- Configurable sample rate and bitrate per codec
- Comfort noise (RFC 3389)
- VAD / silence suppression

### Video Features
- H.264 (AVC)
- H.265 (HEVC)
- VP8
- AV1 (recent versions)
- Video preview / self-view
- Adaptive video bitrate
- Camera selection / hot-switching
- Video bandwidth control
- Video snapshot capture

### Network / Transport
- UDP
- TCP
- TLS (full certificate chain verification)
- DTLS
- SIP over WebSocket (WS / WSS)
- IPv4
- IPv6
- Dual-stack operation
- DNS SRV lookup
- DNS NAPTR lookup
- Multi-homed host handling
- Connection reuse

### NAT Traversal
- STUN (RFC 5389)
- TURN (RFC 5766) with TCP and UDP relay
- ICE full implementation (RFC 5245 / RFC 8445)
- ICE-Lite interoperability
- rport (RFC 3581)
- UPnP (via libupnp)

### Security
- TLS for signaling (TLSv1.0 through TLSv1.3)
- SRTP with SDES key exchange (RFC 4568)
- ZRTP media encryption (via bzrtp library, RFC 6189) with SAS verification
- DTLS-SRTP (RFC 5764)
- Post-quantum ZRTP key exchange (experimental, recent versions)
- Certificate pinning
- Custom CA certificates
- Client certificate authentication (mutual TLS)
- LIME (Linphone Instant Message Encryption) - Double Ratchet E2E encryption

### Presence & BLF
- SIP SUBSCRIBE/NOTIFY presence (RFC 3856)
- SIP PUBLISH for presence state (RFC 3903)
- Rich presence (RFC 4480)
- Buddy / friend list with presence monitoring
- Presence states: online, busy, away, offline, etc.

### Call History / Logs
- Call log: incoming, outgoing, missed, declined
- Call duration
- Per-call quality statistics
- Persistent storage (SQLite database)
- Call log search

### Contact Management
- Local contact database
- vCard import / export
- CardDAV synchronization
- Native / system contacts integration (iOS, Android, macOS)
- LDAP directory (via plugin)
- Contact search / filter
- Contact groups

### DTMF
- RFC 2833 / RFC 4733 (telephone-event)
- SIP INFO method
- Inband DTMF
- Configurable DTMF method

### Voicemail
- MWI via SUBSCRIBE/NOTIFY (RFC 3842)
- Voicemail URI per account

### Multiple Accounts / Lines
- Multiple simultaneous accounts
- Per-account settings (codec, transport, proxy, encryption)
- Default account selection
- Account enable / disable

### Provisioning
- Remote provisioning via XML configuration URL (HTTP/HTTPS)
- QR code provisioning
- Configuration file provisioning
- FlexiSIP server integration

### QoS / DSCP
- DSCP marking for signaling and media
- Adaptive jitter buffer
- Bandwidth estimation and adaptation
- Configurable audio / video bandwidth limits

### Diagnostics
- Full SIP trace logging (belle-sip)
- Verbose debug logging (ortp, mediastreamer2, liblinphone layers)
- Log file output / rotation
- Real-time call statistics (jitter, packet loss, bandwidth, codec)
- RTCP statistics

### Messaging
- SIP MESSAGE (RFC 3428) instant messaging
- Group chat (server-based conference chat rooms)
- File transfer (HTTP upload or SIP MESSAGE)
- LIME end-to-end encrypted messaging (Double Ratchet)
- IMDN delivery / read receipts
- Persistent chat rooms (server-hosted)
- Ephemeral / self-destructing messages

### Advanced / Developer Features
- Full C/C++ SDK (liblinphone) with language bindings:
  - Java / Kotlin (Android)
  - Swift / Objective-C (iOS)
  - C# (.NET)
  - Python
- Acoustic echo cancellation (WebRTC AEC module in mediastreamer2)
- Noise suppression
- Automatic gain control (AGC)
- Audio route selection / switching
- Push notifications (iOS APNs, Android FCM)
- Background mode for persistent registration
- Conference server hosting capability
- Custom SIP header injection via API
- Full SIP event framework access
- Plugin architecture for mediastreamer2 (custom codecs, filters)

---

## 3. Zoiper

**Platform:** Windows, macOS, Linux, iOS, Android
**License:** Free (limited) / Pro / Business (commercial)
**Notable:** Also supports IAX2 protocol alongside SIP

### SIP Registration & Authentication
- Multiple SIP accounts (Free: 1 SIP + 1 IAX; Pro/Business: unlimited)
- IAX2 protocol support alongside SIP
- SIP digest authentication
- Registration expiry configuration
- Outbound proxy
- Auth username / SIP username separation
- Display name
- Transport selection (UDP / TCP / TLS)
- SIP keep-alive (OPTIONS, CRLF, custom interval)

### Call Features
- Basic calling
- Call hold / resume
- Attended transfer (Pro/Business)
- Blind transfer
- 3-way conference (Pro/Business)
- Call waiting
- Do Not Disturb
- Auto-answer with header detection (Pro/Business)
- Call recording (Pro/Business)
- Mute / unmute
- Speaker toggle / audio routing
- Call timer
- Click-to-call URL handlers (sip:, tel:)
- Call pickup (server-dependent)
- Intercom / paging via auto-answer header
- Speed dial
- Redial
- Action URL / HTTP webhooks on call events (Business)
- Call encryption indicator
- Early media
- Call forwarding configuration

### Audio Codec Support
- G.711 u-law (PCMU)
- G.711 A-law (PCMA)
- GSM
- Speex (NB, WB, UWB)
- iLBC
- G.729 (Pro/Business, licensed)
- G.722 (HD Voice)
- Opus (Pro/Business)
- Codec priority ordering
- VAD / silence detection

### Video Features (Pro/Business)
- H.264
- H.263+
- VP8
- Video calls
- Camera selection

### Network / Transport
- UDP
- TCP
- TLS
- IPv4
- IPv6 (Pro/Business)
- DNS SRV
- DNS NAPTR

### NAT Traversal
- STUN
- TURN (Pro/Business)
- ICE (Pro/Business)
- rport

### Security
- TLS for signaling
- SRTP with SDES (Pro/Business)
- ZRTP (Pro/Business)

### Presence & BLF
- BLF - Busy Lamp Field (Pro/Business)
- SIP presence (SUBSCRIBE/NOTIFY)
- BLF panel with speed dial

### Call History / Logs
- Full call history (incoming, outgoing, missed)
- Call duration
- History search

### Contact Management
- Local contacts
- System / OS contacts integration
- Contact search / filter
- LDAP directory (Business)
- Favorites
- Contact import

### DTMF
- RFC 2833 / RFC 4733
- SIP INFO
- Inband DTMF
- Configurable per account

### Voicemail
- MWI (SUBSCRIBE/NOTIFY)
- Voicemail speed-dial button
- Voicemail number per account

### Multiple Accounts / Lines
- Multiple accounts (Free: limited; Pro/Business: unlimited)
- Per-account configuration
- Account enable / disable
- Default account selection

### Provisioning
- Auto-provisioning via HTTPS URL
- Mass deployment tools (Business)
- QR code provisioning (mobile)
- Configuration profiles

### QoS / DSCP
- DSCP / ToS marking
- Adaptive jitter buffer

### Diagnostics
- SIP message trace
- Debug logging
- Log export

### Advanced Features
- IAX2 protocol support (unique among listed softphones)
- Action URLs / HTTP webhooks on call events (Business)
- Command-line dialing
- Custom ringtones (per contact, per account)
- Headset integration (Jabra, Plantronics/Poly, Sennheiser/EPOS)
- Acoustic echo cancellation
- Noise reduction
- Automatic gain control
- Skin / theme customization
- Hotkey / keyboard shortcut configuration
- Popup on incoming call
- CRM integration API (Business)
- Fax T.38 support (Pro)
- Click-to-call browser plugin
- System tray mode

---

## 4. MicroSIP

**Platform:** Windows only
**License:** Open Source (GPLv2)
**SIP Stack:** PJSIP
**Notable:** Ultra-lightweight single portable executable (~3MB)

### SIP Registration & Authentication
- Multiple SIP account support (via INI config)
- SIP digest authentication
- Outbound proxy
- Auth username / SIP username separation
- Display name
- Transport selection (UDP / TCP / TLS)
- Domain / registrar configuration
- Registration expiry configuration

### Call Features
- Basic calling (INVITE / BYE)
- Call hold / resume
- Attended transfer
- Blind transfer
- Conference calling (multi-party via PJSIP conference bridge)
- Call waiting / multiple calls
- Mute / unmute
- Auto-answer (via config or command-line flag)
- Call recording (WAV)
- Do Not Disturb
- Audio device selection (speaker / microphone)
- Click-to-call via command-line
- URL handler registration (sip:, tel:)
- Redial
- Speed dial panel (combined with BLF)
- Call forwarding (server-dependent)

### Audio Codec Support
- G.711 u-law (PCMU)
- G.711 A-law (PCMA)
- G.722 (wideband)
- G.729 (when compiled with support)
- GSM
- Speex (NB, WB)
- iLBC
- Opus
- SILK
- L16 (uncompressed linear PCM)
- Codec priority ordering

### Video Features
- No video support (audio-only by design for lightweight footprint)

### Network / Transport
- UDP
- TCP
- TLS
- IPv4
- IPv6 (via PJSIP)
- DNS SRV lookup

### NAT Traversal
- STUN (via PJSIP)
- TURN (via PJSIP)
- ICE (via PJSIP)
- rport

### Security
- TLS for signaling
- SRTP with SDES (via PJSIP)

### Presence & BLF
- BLF (Busy Lamp Field) with visual speed-dial panel
- Presence indication on BLF buttons (dialog event package)
- BLF panel is a prominent, well-regarded feature

### Call History / Logs
- Call history (incoming, outgoing, missed)
- Missed call notification
- History stored locally

### Contact Management
- Speed-dial / BLF contact panel
- Contacts defined in INI config file
- Minimal contact management (lightweight design philosophy)

### DTMF
- RFC 2833 / RFC 4733
- SIP INFO
- Inband DTMF
- Configurable DTMF method

### Voicemail
- Basic MWI support
- Voicemail speed-dial button

### Multiple Accounts / Lines
- Multiple accounts via config
- Account switching

### Provisioning
- INI file-based configuration (microsip.ini)
- Portable mode (no installation, no registry, run from USB)
- Registry-based settings (optional)
- Trivial mass deployment (copy exe + ini file)

### QoS / DSCP
- DSCP marking (via PJSIP)

### Diagnostics
- SIP message logging
- Debug log output

### Advanced Features
- Ultra-lightweight (~3MB single portable EXE)
- Portable mode (run from USB, no install, no registry writes)
- Extensive command-line interface:
  - `/call:number` - initiate call
  - `/hangupall` - hang up all calls
  - `/minimize` - minimize to tray
  - `/answer` - answer incoming call
  - `/transfer:number` - transfer active call
  - `/dtmf:digits` - send DTMF
  - `/hold` - hold/resume
  - `/conference` - merge calls
- System tray operation
- Minimal resource usage (very low RAM/CPU footprint)
- Built on PJSIP (inherits PJSIP capabilities)
- Custom ringtones (WAV files)
- Windows toast notifications
- Easy scripting / automation via command-line
- Simple INI config format

---

## 5. Other Notable SIP Clients

Note: No well-known SIP softphone named "Obelisk" was identified.
Below are other notable SIP clients for competitive reference.

### 5a. Jitsi
- Open source (Apache 2.0), cross-platform
- SIP + XMPP dual protocol support
- Multiple account registration
- Audio + video calling (H.264, VP8)
- Call transfer (attended + blind), conference
- SRTP (SDES), ZRTP, TLS
- ICE, STUN, TURN for NAT traversal
- SIP MESSAGE instant messaging + OTR encryption
- Call recording
- Desktop sharing / screen sharing

### 5b. Baresip
- Lightweight modular SIP user agent in C
- Open source (BSD license)
- Based on libre (SIP) + librem (media) + baresip (UA)
- Modular architecture: codec modules, UI modules, audio driver modules
- Console / headless operation (ideal for embedded / CLI)
- All standard codecs via modules (G.711, G.722, Opus, etc.)
- ICE, STUN, TURN
- SRTP, DTLS-SRTP, TLS
- Multiple accounts
- Call transfer, hold, conference
- Very low resource usage
- Excellent for automation and scripting

### 5c. PhonerLite
- Freeware Windows SIP softphone

### 5d. Twinkle
- Open source Linux SIP softphone (Qt)
- Full-featured: transfer, conference, call scripts
- ZRTP support

### 5e. Ekiga (formerly GnomeMeeting)
- Open source, Linux/Windows
- SIP + H.323 dual protocol
- Video calling (H.264, H.261, H.263)
- LDAP directory

### 5g. Groundwire (by CounterPath)
- Mobile-focused premium SIP client (iOS/Android)
- Push notification support for incoming calls
- All Bria-like features in mobile form factor
- Background persistent registration via push
- Enterprise provisioning

---

## 6. Consolidated Feature Matrix for siptty Reference

This section synthesizes ALL features across the above clients into a
single master list, categorized for use as a reference when building
siptty.

### A. SIP Registration & Account Management
1. SIP REGISTER with digest auth (MD5, SHA-256)
2. Multiple simultaneous account registrations
3. Auth username separate from SIP URI user
4. Display name per account
5. Outbound proxy per account
6. Transport per account (UDP, TCP, TLS, WSS)
7. Registration expiry / refresh configuration
8. SIP keep-alive (CRLF, OPTIONS, re-REGISTER)
9. Registrar failover / redundancy
10. DNS SRV / NAPTR / A record resolution
11. SIP Outbound (RFC 5626) with flow tokens
12. GRUU (RFC 5627)
13. Account enable/disable without deletion
14. Default account selection
15. Registration retry with exponential backoff

### B. Call Control
1. Make call (INVITE)
2. Answer call (200 OK)
3. Reject call (486/603)
4. Decline with reason
5. Hang up (BYE)
6. Cancel outgoing (CANCEL)
7. Call hold / resume (re-INVITE sendonly/recvonly)
8. Blind transfer (REFER)
9. Attended transfer (REFER with Replaces)
10. 3-way / N-way local conference (audio mixing)
11. Server-side conference (focus URI)
12. Call waiting / multiple concurrent calls
13. Call forwarding (302 redirect, server CFB/CFNA/CFU)
14. Do Not Disturb (local)
15. Auto-answer (configurable, header-triggered: Alert-Info, Call-Info)
16. Call recording (local WAV)
17. Mute / unmute
18. Volume control (speaker + microphone)
19. Call timer display
20. Early media (183 with SDP)
21. 100rel / PRACK (RFC 3262)
22. Session timers (RFC 4028)
23. UPDATE method (RFC 3311)
24. Re-INVITE for renegotiation
25. Intercom / paging (auto-answer with header)
26. Call park / pickup (server-dependent)
27. Speed dial
28. Redial
29. Click-to-call (sip: / tel: URI handler)

### C. Audio Codecs
1. G.711 u-law (PCMU, PT 0)
2. G.711 A-law (PCMA, PT 8)
3. G.722 (PT 9, wideband)
4. Opus (dynamic PT, NB-FB, variable bitrate)
5. G.729 / G.729a (PT 18)
6. iLBC (dynamic PT)
7. GSM-FR (PT 3)
8. Speex (NB/WB/UWB, dynamic PT)
9. SILK (dynamic PT)
10. L16 (dynamic PT, uncompressed)
11. BV16 (dynamic PT)
12. AAC-ELD (dynamic PT)
13. Comfort noise (CN, PT 13, RFC 3389)
14. Telephone-event (PT 101 typ., RFC 4733)
15. Codec priority ordering
16. VAD / silence suppression
17. Packet loss concealment (PLC)
18. Adaptive bitrate (Opus, Speex)

### D. DTMF
1. RFC 2833 / RFC 4733 (telephone-event in RTP)
2. SIP INFO (application/dtmf-relay, application/dtmf)
3. Inband DTMF (audio tones in RTP stream)
4. Configurable method per account
5. Dial pad UI for mid-call DTMF

### E. Network & Transport
1. UDP transport
2. TCP transport
3. TLS transport (with cert validation, SNI)
4. DTLS
5. SIP over WebSocket (WS / WSS, RFC 7118)
6. IPv4
7. IPv6
8. Dual-stack
9. DNS SRV (RFC 3263)
10. DNS NAPTR (RFC 3263)
11. A / AAAA fallback
12. Multi-homed host handling
13. Connection reuse
14. Keep-alive mechanisms

### F. NAT Traversal
1. STUN (RFC 5389)
2. TURN (RFC 5766) - UDP and TCP relay
3. ICE (RFC 5245 / RFC 8445) - full implementation
4. ICE-Lite interop
5. rport (RFC 3581)
6. Symmetric RTP
7. UPnP IGD
8. Outbound proxy for NAT
9. External/public IP override

### G. Security & Encryption
1. TLS for SIP signaling (TLSv1.2, TLSv1.3)
2. SRTP with SDES key exchange (RFC 3711, RFC 4568)
3. ZRTP (RFC 6189) with SAS verification
4. DTLS-SRTP (RFC 5764)
5. Mutual TLS (client certificates)
6. CA certificate management / custom bundles
7. Certificate hostname verification
8. Certificate pinning
9. Encrypted credential storage
10. E2E messaging encryption (LIME/Double Ratchet)

### H. Presence & BLF
1. SIP SUBSCRIBE/NOTIFY presence (RFC 3856)
2. SIP PUBLISH (RFC 3903)
3. Busy Lamp Field / dialog event package (RFC 4235)
4. Presence states (online, busy, away, on-phone, offline)
5. Custom status messages
6. Rich presence (RFC 4480)
7. BLF + speed-dial combined panel

### I. Voicemail
1. MWI - Message Waiting Indicator (RFC 3842, SUBSCRIBE/NOTIFY)
2. Visual voicemail indicator (icon/badge)
3. One-touch voicemail dial
4. Voicemail number per account

### J. Call History & Logs
1. Call log: incoming, outgoing, missed, declined
2. Call duration per entry
3. Call quality stats per entry
4. Search / filter history
5. Persistent storage
6. Export capability

### K. Contact Management
1. Local contact database
2. LDAP directory integration
3. CardDAV synchronization
4. vCard import / export
5. System/OS contacts integration
6. Contact search / filter
7. Contact groups
8. Favorites
9. Corporate directory lookup

### L. Messaging
1. SIP MESSAGE (RFC 3428) instant messaging
2. Group chat
3. File transfer
4. Delivery / read receipts (IMDN)
5. E2E encrypted messaging
6. SMS via SIP MESSAGE

### M. QoS & Media Quality
1. DSCP / ToS marking for SIP packets
2. DSCP / ToS marking for RTP packets
3. Separate DSCP for audio vs video
4. Adaptive jitter buffer
5. Bandwidth estimation / adaptation
6. Packet loss concealment (PLC)

### N. Audio Processing
1. Acoustic echo cancellation (AEC)
2. Noise suppression / reduction
3. Automatic gain control (AGC)
4. Comfort noise generation
5. VAD / silence suppression

### O. Provisioning & Deployment
1. Configuration file (INI, XML, JSON)
2. Remote provisioning via URL (HTTP/HTTPS)
3. QR code provisioning
4. Zero-touch / auto-provisioning
5. Mass deployment support
6. Portable mode (no install)
7. Push notification registration (mobile)

### P. Diagnostics & Debugging
1. Full SIP message trace
2. Configurable log levels (debug/info/warn/error)
3. Log file output / rotation / export
4. Real-time call quality stats (jitter, loss, RTT, MOS)
5. RTCP statistics
6. Audio device diagnostics
7. Network quality indicators

### Q. UI & Integration
1. System tray / notification area
2. Click-to-call URL handler
3. Command-line interface / arguments
4. Headset integration (Jabra, Plantronics, Sennheiser)
5. Custom ringtones (per contact, per account)
6. Hotkey / keyboard shortcut configuration
7. CRM integration API
8. Action URLs / webhooks on call events
9. Desktop notifications (toast/popup)
10. Skin / theme customization

### R. Protocol Extensions (Advanced)
1. IAX2 protocol support (Zoiper)
2. XMPP protocol support (Bria, Jitsi)
3. H.323 protocol support (Ekiga)
4. T.38 fax support (Zoiper Pro)
5. SIP over WebSocket for WebRTC interop

---

## 7. Key Takeaways for siptty

### Must-Have (Core) Features
Every serious SIP client implements these:
- SIP REGISTER with digest auth
- Basic call control (INVITE/BYE/CANCEL/hold/resume)
- At least G.711 u-law + A-law codecs
- RFC 4733 DTMF (telephone-event)
- UDP + TCP transport
- STUN for NAT traversal
- Call history
- Multiple account support
- Mute/unmute
- Volume control

### Expected Features (Standard)
Most commercial/mature clients have:
- TLS signaling
- SRTP media encryption (SDES)
- G.722, Opus codecs
- ICE/TURN NAT traversal
- Blind + attended transfer
- 3-way conference
- BLF / presence
- MWI (voicemail indicator)
- Do Not Disturb
- Auto-answer
- DNS SRV resolution
- Call recording
- DSCP/QoS marking
- SIP message logging / diagnostics

### Differentiating Features (Nice-to-Have)
These set premium clients apart:
- ZRTP encryption
- DTLS-SRTP
- Video calling
- WebSocket transport
- IPv6
- LDAP/CardDAV contacts
- Push notifications
- CRM integration / Action URLs
- Headset SDK integration
- Advanced provisioning
- Command-line automation
- Plugin/module architecture

### Design Insights for a Terminal/CLI SIP Client
- MicroSIP's command-line interface is the closest analog
- Baresip is the gold standard for headless/CLI SIP operation
- Key CLI operations: /call, /hangup, /answer, /transfer, /hold, /dtmf, /mute
- INI/TOML config file is the simplest provisioning approach
- SIP trace logging is essential for debugging
- PJSIP is the most popular underlying stack for building new clients
