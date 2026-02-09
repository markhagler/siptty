# SIP RFCs & Standards Reference

Comprehensive list of RFCs relevant to implementing a full-featured SIP User Agent.

---

## 1. Core SIP Protocol

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3261 | SIP: Session Initiation Protocol | The base SIP spec — defines INVITE, ACK, BYE, CANCEL, REGISTER, OPTIONS, transaction layer, dialog management, and UA behavior. |
| RFC 3263 | Locating SIP Servers | DNS resolution procedures: NAPTR → SRV → A/AAAA for finding SIP servers and selecting transport. |
| RFC 3264 | An Offer/Answer Model with SDP | How SDP offers and answers are exchanged to negotiate media sessions. |
| RFC 6665 | SIP-Specific Event Notification | The SUBSCRIBE/NOTIFY framework for event subscriptions (replaces RFC 3265). |
| RFC 3261 §17 | SIP Transactions | Client and server transaction state machines, retransmission timers (T1, T2, Timer A–K). |
| RFC 3261 §12 | SIP Dialogs | Dialog creation, route set management, remote target updates. |

## 2. SIP Method Extensions

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3262 | Reliability of Provisional Responses in SIP | PRACK method + `100rel` option tag — makes 1xx responses (except 100) reliable. |
| RFC 3311 | The SIP UPDATE Method | Modify session parameters before dialog confirmation without a full re-INVITE. |
| RFC 3515 | The SIP REFER Method | Instructs a UA to send a request to a third party — the basis for call transfer. |
| RFC 3428 | SIP Extension for Instant Messaging | MESSAGE method for sending instant messages within or outside a dialog. |
| RFC 6086 | Session Initiation Protocol (SIP) INFO Method and Info Package Framework | Revised INFO method for carrying application-level data (e.g., DTMF) within a dialog. |
| RFC 3903 | SIP Extension for Event State Publication | PUBLISH method — allows a UA to publish event state (e.g., presence) to an Event State Compositor. |
| RFC 3265 | SIP-Specific Event Notification (original) | Original SUBSCRIBE/NOTIFY spec — obsoleted by RFC 6665 but widely referenced. |

## 3. Registration

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3261 §10 | REGISTER Method | Core registration — binds an AOR (address-of-record) to a contact URI at the registrar. |
| RFC 5626 | Managing Client-Initiated Connections in SIP (Outbound) | Persistent outbound connections through NAT — UA maintains a flow to the edge proxy. |
| RFC 5627 | Obtaining and Using GRUUs in SIP | Globally Routable UA URIs — a URI that routes to a specific UA instance, survives re-registration. |
| RFC 3327 | SIP Extension Header Field for Registering Non-Adjacent Contacts (Path) | Path header lets proxies insert themselves during registration for future request routing. |
| RFC 3608 | SIP Extension Header Field for Service Route Discovery | Service-Route header returned during registration tells UA which outbound proxy route to use. |
| RFC 6140 | Registration for Multiple Phone Numbers in SIP | Register multiple AORs in a single REGISTER transaction. |

## 4. Session & Media Negotiation

| RFC | Title | Description |
|------|-------|-------------|
| RFC 4566 | SDP: Session Description Protocol | The format for describing multimedia sessions — codecs, IP addresses, ports, media attributes. |
| RFC 3264 | An Offer/Answer Model with SDP | Negotiation model: one side offers SDP capabilities, the other answers with selected parameters. |
| RFC 6337 | SIP Usage of the Offer/Answer Model | Clarifies when offers/answers occur in INVITE, re-INVITE, UPDATE, PRACK flows. |
| RFC 4028 | Session Timers in SIP | Periodic re-INVITE or UPDATE to keep sessions alive and detect dead endpoints (`Session-Expires` header). |
| RFC 3312 | Integration of Resource Management and SIP | Preconditions — delay session establishment until required resources (QoS) are reserved. |
| RFC 3840 | Indicating UA Capabilities in SIP | Feature tags in Contact headers to advertise capabilities (audio, video, methods). |
| RFC 3841 | Caller Preferences for SIP | Caller expresses routing preferences based on callee UA capabilities. |
| RFC 5939 | SDP Capability Negotiation | Framework for negotiating potential SDP configurations before committing. |

## 5. NAT Traversal

| RFC | Title | Description |
|------|-------|-------------|
| RFC 8489 | STUN: Session Traversal Utilities for NAT | Discover public IP:port mappings; used for ICE connectivity checks (updates RFC 5389). |
| RFC 5389 | STUN (original) | Original STUN spec — still widely referenced. |
| RFC 8656 | TURN: Traversal Using Relays around NAT | Relay server for media when direct P2P connectivity through NAT fails (updates RFC 5766). |
| RFC 5766 | TURN (original) | Original TURN spec. |
| RFC 8445 | ICE: Interactive Connectivity Establishment | Framework using STUN+TURN to find the best media path through NATs via connectivity checks (updates RFC 5245). |
| RFC 5245 | ICE (original) | Original ICE spec. |
| RFC 8838 | Trickle ICE | Exchange ICE candidates incrementally to reduce call setup time. |
| RFC 3581 | Symmetric Response Routing (rport) | `rport` in Via header — responses sent to NAT-mapped source IP:port. |
| RFC 5626 | SIP Outbound | Persistent outbound connections through NAT for signaling. |
| RFC 5768 | Indicating Support for ICE in SIP | SDP attributes and SIP mechanisms to signal ICE support. |
| RFC 7362 | Latching: Hosted NAT Traversal (HNT) | Media-aware proxy sends media to the address it receives packets from. |

## 6. Security

| RFC | Title | Description |
|------|-------|-------------|
| RFC 8446 | TLS 1.3 | Transport Layer Security — encrypts SIP signaling (sips: URIs, TLS transport). |
| RFC 5246 | TLS 1.2 | Previous TLS version, still widely deployed for SIP. |
| RFC 3711 | The Secure Real-time Transport Protocol (SRTP) | Encryption, authentication, and replay protection for RTP/RTCP media streams. |
| RFC 4568 | SDP Security Descriptions (SDES) | `a=crypto` in SDP carries SRTP keying material — simple but requires secure signaling. |
| RFC 5763 | Framework for Establishing a DTLS-SRTP Context | DTLS-SRTP framework for key exchange in the media path (used by WebRTC). |
| RFC 5764 | DTLS Extension to Establish Keys for SRTP | DTLS-SRTP transport — `a=fingerprint` and `a=setup` in SDP. |
| RFC 6189 | ZRTP: Media Path Key Agreement for Unicast Secure RTP | DH key exchange in media path — no PKI needed, uses Short Authentication String (SAS) for voice verification. |
| RFC 3830 | MIKEY: Multimedia Internet KEYing | Key management protocol for SRTP transportable in SDP. |
| RFC 8224 | Authenticated Identity Management in SIP | Updated SIP Identity — uses PASSporT tokens to sign caller identity (STIR). |
| RFC 8225 | PASSporT: Personal Assertion Token | JSON token for cryptographically signed caller identity assertions (STIR/SHAKEN). |
| RFC 8226 | Secure Telephone Identity Credentials | Certificate framework for STIR/SHAKEN telephone identity verification. |
| RFC 3329 | Security Mechanism Agreement for SIP | UA and proxy negotiate which security mechanism to use (TLS, IPsec, Digest). |
| RFC 3323 | A Privacy Mechanism for SIP | Privacy header — request that identity be hidden from the remote party. |
| RFC 3325 | P-Asserted-Identity | Convey authenticated caller identity within a trusted network. |

## 7. Authentication

| RFC | Title | Description |
|------|-------|-------------|
| RFC 2617 | HTTP Authentication: Basic and Digest | Original Digest auth used by SIP for 401/407 challenges (MD5-based). |
| RFC 7616 | HTTP Digest Access Authentication | Updated Digest auth — adds SHA-256, userhash, improvements over RFC 2617. |
| RFC 8760 | Using Digest with SHA-256/SHA-512-256 in SIP | SIP-specific guidance on modern hash algorithms for Digest auth. |
| RFC 3261 §22 | SIP Authentication | SIP's use of Digest auth — UA responds to 401 (UAS) or 407 (proxy) challenges. |

## 8. DTMF Signaling

| RFC | Title | Description |
|------|-------|-------------|
| RFC 4733 | RTP Payload for DTMF Digits, Telephony Tones, and Signals | `telephone-event` RTP payload for out-of-band DTMF — the standard method (obsoletes RFC 2833). |
| RFC 2833 | RTP Payload for DTMF (original) | Original DTMF-over-RTP spec — superseded by RFC 4733 but still widely referenced. |
| RFC 6086 | SIP INFO for DTMF | INFO method carries `application/dtmf-relay` — used by some legacy systems. |
| RFC 4730 | KPML: Key Press Stimulus Event Package | Subscribe to key press events for advanced IVR stimulus signaling. |

## 9. Supplementary Services (Call Features)

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3515 | The SIP REFER Method | Instruct a UA to send a request to a third party — blind and attended call transfer. |
| RFC 3891 | The SIP "Replaces" Header Field | INVITE replaces an existing dialog — key mechanism for attended (consultative) transfer. |
| RFC 3892 | The SIP Referred-By Mechanism | `Referred-By` header tells transfer target who initiated the REFER. |
| RFC 3911 | The SIP "Join" Header Field | INVITE joins an existing dialog — barge-in to calls. |
| RFC 4579 | SIP Call Control - Conferencing for UAs | UA behavior for creating, joining, and managing conference calls. |
| RFC 4575 | A SIP Event Package for Conference State | Subscribe to conference events — participant list, join/leave notifications. |
| RFC 3326 | The Reason Header Field for SIP | Reason header in BYE/CANCEL — carries disconnect cause (Q.850 codes). |
| RFC 4244 | Request History Information | History-Info header — tracks request path through diversions/redirections. |
| — | Call Hold | Re-INVITE with `a=sendonly`/`a=inactive` in SDP — no separate RFC, uses SDP direction attributes. |
| — | Music on Hold | `a=sendonly` + local media source, or REFER to MoH server — implementation-specific. |

## 10. Presence & Event Packages

| RFC | Title | Description |
|------|-------|-------------|
| RFC 6665 | SIP-Specific Event Notification | Core SUBSCRIBE/NOTIFY framework for all event packages. |
| RFC 3856 | A Presence Event Package for SIP | "presence" event package — subscribe to online/offline/busy/away status. |
| RFC 3863 | Presence Information Data Format (PIDF) | XML format for presence information — status, contact, notes, timestamps. |
| RFC 4480 | RPID: Rich Presence Extensions to PIDF | Rich presence — activities (on-the-phone, meeting), mood, place. |
| RFC 4482 | CIPID: Contact Information for PIDF | Extends PIDF with contact info — display name, icon URL, homepage. |
| RFC 3842 | Message Summary and MWI Event Package | Voicemail/message waiting indication — the envelope icon on your phone. |
| RFC 4235 | An INVITE-Initiated Dialog Event Package | Dialog state events — monitor call state of other users; basis for BLF (Busy Lamp Field). |
| RFC 3903 | SIP PUBLISH Method | Publish own event state (e.g., presence) to an event state compositor. |
| RFC 4662 | SIP Event Notification Extension for Resource Lists (RLS) | Subscribe to a list of resources (buddy list) with a single SUBSCRIBE. |
| RFC 5262 | PIDF Extension for Partial Presence | Partial notifications — only send changed presence data. |
| RFC 4660 | Event Notification Filtering | Subscribers filter notifications to reduce bandwidth. |
| RFC 3680 | A SIP Event Package for Registrations | Subscribe to registration state changes at a registrar. |
| RFC 3857 | Watcher Information Event Template-Package | Lets a presentity know who is watching their presence. |

## 11. Media / RTP / Codecs

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3550 | RTP: A Transport Protocol for Real-Time Applications | Core RTP — packet format, sequence numbers, timestamps, SSRC, payload types, RTCP. |
| RFC 3551 | RTP Profile for Audio and Video Conferences (AVP) | Standard RTP/AVP profile — payload type numbers (PCMU=0, PCMA=8, G722=9, G729=18). |
| RFC 3711 | SRTP | Adds encryption (AES-CM), authentication (HMAC-SHA1), replay protection to RTP. |
| RFC 3389 | RTP Payload for Comfort Noise | Comfort Noise (CN) payload — generates background noise during silence in VAD/DTX. |
| RFC 5761 | Multiplexing RTP and RTCP on a Single Port | `a=rtcp-mux` — RTP and RTCP share the same port, simplifies NAT traversal. |
| RFC 3605 | RTCP Attribute in SDP | `a=rtcp:` attribute to specify a separate RTCP port if not RTP+1. |
| RFC 4585 | Extended RTP Profile for RTCP-Based Feedback (AVPF) | Immediate RTCP feedback — PLI, FIR, NACK for video. |
| RFC 5124 | Extended Secure RTP Profile for RTCP-Based Feedback (SAVPF) | SRTP + RTCP feedback combined — the profile used by WebRTC. |
| RFC 7587 | RTP Payload Format for Opus | Payload format for Opus codec — 6–510 kbps, narrowband to fullband. |
| RFC 6716 | Definition of the Opus Audio Codec | Opus codec specification — excellent quality at all bitrates, voice and music. |
| RFC 3952 | RTP Payload Format for iLBC | Payload format for iLBC codec — 13.33 or 15.2 kbps. |
| RFC 4867 | RTP Payload Format for AMR and AMR-WB | AMR (narrowband) and AMR-WB (wideband/HD) codecs for mobile networks. |
| RFC 4733 | RTP Payload for DTMF/Telephony Events | `telephone-event` payload for DTMF digits over RTP. |
| ITU-T G.711 | PCM of Voice Frequencies | PCMU (μ-law) and PCMA (A-law) — fundamental 64 kbps codecs, universally supported. |
| ITU-T G.722 | 7 kHz Audio within 64 kbit/s | Wideband (HD Voice) at 64 kbps — much better quality than G.711. |
| ITU-T G.729 | Speech Coding at 8 kbit/s (CS-ACELP) | Low-bitrate codec (8 kbps) — widely used, patents expired 2017. |

## 12. DNS Resolution

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3263 | Locating SIP Servers | Complete DNS procedure: NAPTR → SRV → A/AAAA with transport selection. |
| RFC 2782 | DNS SRV RR | SRV records — `_sip._udp.example.com` gives priority, weight, port, target for SIP. |
| RFC 3403 | DNS NAPTR RR | NAPTR records — determine available transports (UDP, TCP, TLS) and preference. |
| RFC 7984 | Locating SIP Servers in Dual-Stack Networks | Updated DNS procedures for IPv4/IPv6 dual-stack environments. |

## 13. Reliability & Timers

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3262 | Reliability of Provisional Responses (100rel/PRACK) | Makes 1xx responses reliable — UAS retransmits until UAC sends PRACK. |
| RFC 4028 | Session Timers in SIP | Periodic session refresh via re-INVITE/UPDATE to detect failed endpoints and prevent zombie sessions. |
| RFC 5765 | SIP Transaction Timers | T1 (500ms RTT), T2 (4s max retransmit), Timer B (32s INVITE timeout) — defined in RFC 3261 §17. |
| RFC 5923 | Connection Reuse in SIP | Reuse existing TCP/TLS connections for subsequent transactions. |

## 14. WebSocket Transport

| RFC | Title | Description |
|------|-------|-------------|
| RFC 7118 | The WebSocket Protocol as a Transport for SIP | SIP over WebSocket (ws:/wss:) — enables browser-based SIP UAs / WebRTC-to-SIP gateways. |
| RFC 6455 | The WebSocket Protocol | Base WebSocket protocol used as transport substrate. |

## 15. Interoperability Headers

| RFC | Title | Description |
|------|-------|-------------|
| RFC 3325 | P-Asserted-Identity / P-Preferred-Identity | Convey authenticated caller identity in trusted networks. |
| RFC 3323 | Privacy Mechanism for SIP | Privacy header — request identity hiding from remote party. |
| RFC 3455 | Private Header Extensions for 3GPP | P-Access-Network-Info, P-Visited-Network-ID, P-Charging-Vector for IMS. |
| RFC 3326 | Reason Header Field | Disconnect cause in BYE/CANCEL — includes Q.850 cause codes. |
| RFC 4244 | Request History Information | History-Info header — tracks diversions/redirections. |
| RFC 3327 | Path Header | Proxies insert themselves during registration for routing. |
| RFC 3608 | Service-Route Header | Returned during registration — tells UA the outbound route. |

---

## Implementation Priority

### Tier 1 — Minimum Viable SIP UA
- RFC 3261 — Core SIP
- RFC 3264 + RFC 4566 — SDP Offer/Answer
- RFC 2617/7616 — Digest Authentication
- RFC 3263 — DNS SRV/NAPTR
- RFC 3550/3551 — RTP with AVP profile
- RFC 4733 — DTMF telephone-event
- RFC 3581 — rport
- G.711 PCMU/PCMA codec

### Tier 2 — Standard SIP Phone
- RFC 3262 — 100rel / PRACK
- RFC 3311 — UPDATE
- RFC 4028 — Session Timers
- RFC 3515 + RFC 3891 — REFER + Replaces (call transfer)
- RFC 5389/8445 — STUN / ICE
- RFC 3711 + RFC 4568 — SRTP + SDES
- RFC 5626 — SIP Outbound
- RFC 3842 — MWI
- RFC 5761 — rtcp-mux
- Opus / G.722 codecs

### Tier 3 — Full-Featured UA
- RFC 5627 — GRUU
- RFC 3856/3863 — Presence + PIDF
- RFC 4235 — Dialog events / BLF
- RFC 3903 — PUBLISH
- RFC 3428 — MESSAGE (IM)
- RFC 6189 — ZRTP
- RFC 5766/8656 — TURN
- RFC 4579/4575 — Conferencing
- RFC 7118 — WebSocket transport
- RFC 8224-8226 — STIR/SHAKEN
