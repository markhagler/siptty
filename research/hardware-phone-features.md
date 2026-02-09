# SIP Hardware Desk Phone Feature Research

Comprehensive SIP UA feature inventory across major enterprise IP phone platforms.

---

## 1. Yealink T5x / T4x Series (T54W, T48U, T46U, T43U, T42U, T41U)

### SIP Registration & Accounts
- Up to 16 SIP accounts (T54W/T48U/T46U), 12 (T43U), 6 (T42U), 6 (T41U)
- Simultaneous registration on all accounts
- Per-account outbound proxy, transport, codec, DTMF settings
- SIP server redundancy (failover) per account — primary/secondary server
- DNS SRV / A record / NAPTR lookup for server resolution
- Registration expiry configurable per line
- Server failover with configurable retry intervals

### Call Handling
- Hold / Resume (RFC 3264 re-INVITE with sendonly/recvonly)
- Music on Hold (local MOH file or server-side)
- Blind (unattended) transfer — SIP REFER
- Attended (consultative) transfer — REFER with Replaces
- Semi-attended transfer (transfer while ringing)
- Local 3-way conference (built-in mixer)
- Local 5-way conference (T54W/T48U/T46U models)
- Network conference (server-side via REFER)
- Call park / call retrieve (via feature codes or BLF)
- Directed call pickup (SIP SUBSCRIBE/NOTIFY + INVITE)
- Group call pickup
- Call completion on busy (CCBS) / call completion on no reply (CCNR) — RFC 6910
- Call join / barge-in
- Anonymous call (caller ID blocking — Privacy: id header)
- Anonymous call rejection
- Call recording trigger (via SIP INFO or server feature code)

### Line Keys / BLF
- Programmable line/DSS keys: 27 (T54W), 16 (T48U/T46U), 21 (T43U)
- Multiple key pages (virtual multi-page DSS keys, up to 3 pages on color models)
- BLF (Busy Lamp Field) via SIP SUBSCRIBE to dialog event package (RFC 4235)
- BLF + speed dial combined
- BLF pickup (press BLF key to pick up ringing call)
- BLF Park / BLF Call Park orbit monitoring
- Key types: Line, BLF, Speed Dial, DTMF, Prefix, Local Group, XML Group, LDAP, Conference, Forward, Transfer, Hold, DND, Recall, SMS, Intercom, Directory, Paging, XML Browser, Hot Desking, URL, Phone Lock, ACD, Multicast Paging
- Expansion modules: EXP43 (20 keys × 3 pages), EXP50 (color LCD, 20 keys × 3 pages) — up to 3 modules daisy-chained = 180 additional keys

### Shared Line Appearance (SLA)
- SCA (Shared Call Appearance) — Broadsoft model
- BLA (Bridged Line Appearance) — draft-location / draft-location-conveyance
- Private hold per shared line
- Visual status indicators (idle, active, held, ringing) on shared line keys

### Presence
- SIP SUBSCRIBE/NOTIFY presence (RFC 3856)
- BLF-based presence (dialog event package)
- UC presence integration (with supported servers)

### Intercom / Paging
- Intercom call with auto-answer (sends Alert-Info: info=alert-autoanswer or Call-Info: answer-after=0)
- Configurable auto-answer tone (beep or silent)
- Supports incoming auto-answer for intercom via Alert-Info / Call-Info headers
- Push-to-talk / intercom mode

### Call Forwarding
- Unconditional (always)
- On Busy
- On No Answer (configurable ring timer)
- Per-account or global forwarding rules
- Forward to voicemail

### DND (Do Not Disturb)
- DND on/off (per account or global)
- DND with authorized contacts exception list
- Server-side or phone-side DND
- DND with feature key sync (SUBSCRIBE to feature-event package)

### Call Waiting
- Call waiting tone (configurable)
- Call waiting on/off
- Visual + audio notification of waiting call

### Auto-Answer
- Auto-answer per account or via SIP headers
- Supported headers: `Alert-Info: info=alert-autoanswer`, `Alert-Info: Auto Answer`, `Call-Info: answer-after=0`, `Call-Info: ;answer-after=0`
- Auto-answer with muted microphone option
- Auto-answer delay timer

### DTMF
- RFC 2833 (RTP telephone-event)
- SIP INFO (application/dtmf-relay)
- SIP INFO (application/dtmf)
- In-band audio (G.711 only)
- Configurable per account
- DTMF payload type configurable

### Codec Support
- G.711 µ-law (PCMU)
- G.711 A-law (PCMA)
- G.722 (HD Voice)
- G.729A/B
- G.726-16/24/32/40
- iLBC (20ms / 30ms)
- Opus (T54W and newer firmware)
- Codec priority configurable per account
- VAD (Voice Activity Detection) / CNG (Comfort Noise Generation)
- Acoustic Echo Cancellation (AEC)
- Noise Reduction / Jitter buffer (adaptive)

### Transport & Security
- UDP, TCP, TLS (TLSv1.0/1.1/1.2/1.3 on newer firmware)
- SRTP (SDES key exchange — RFC 4568)
- Optional / mandatory / disabled SRTP per account
- HTTPS for provisioning and web UI
- 802.1x EAP-MD5 / EAP-TLS / EAP-PEAP
- Certificate management (CA cert import, client cert)
- Configurable TLS cipher suites
- Mutual TLS authentication

### NAT Traversal
- STUN (RFC 5389) client
- ICE (on some firmware versions, limited)
- Outbound proxy
- SIP `rport` parameter (RFC 3581)
- Symmetric RTP
- NAT keep-alive (SIP OPTIONS or UDP keep-alive)
- Static NAT / manual IP mapping

### Provisioning
- Auto-provisioning via TFTP, HTTP, HTTPS, FTP
- Yealink RPS (Redirect & Provisioning Service) — cloud zero-touch
- PnP (Plug and Play) via SIP SUBSCRIBE to multicast
- DHCP option 66/43/160 for provisioning URL
- TR-069 (CWMP) support
- Scheduled auto-provisioning (weekly)
- Provisioning via config files (.cfg) — MAC-specific or common
- Boot-time and periodic provisioning
- AES-encrypted config files

### MWI (Message Waiting Indicator)
- SIP SUBSCRIBE/NOTIFY for MWI (RFC 3842)
- Visual indicator (LED) + on-screen icon
- Stutter dial tone for voicemail
- Voicemail speed dial key
- Configurable voicemail number per account

### Call History
- Local call log: placed, received, missed, forwarded
- Typically 100 entries per category
- Call log accessible from phone UI
- Network/server-side call log (BroadSoft XSI, etc.)

### Phonebook / Directory
- Local phonebook: up to 1000 contacts
- Remote XML phonebook (HTTP/HTTPS) — up to 5 remote phonebook URLs
- LDAP directory (configurable search attributes, LDAP/LDAPS)
- BroadSoft XSI directory integration
- Favorites / speed dial list
- Blacklist (call block list)
- Phonebook search across all sources
- Import/export contacts via CSV/XML

### QoS
- DSCP markings configurable for SIP signaling and RTP separately
- Default: SIP=26 (AF31), RTP=46 (EF)
- 802.1p priority tagging (Layer 2)
- LLDP-MED based QoS policy (auto-configure from switch)

### Hotdesking
- Hotdesking via programmable key (clears registration, prompts for new credentials)
- ACD (Automatic Call Distribution) agent login/logout
- Server-controlled hotdesking

### Action URLs
- HTTP/HTTPS GET callbacks on phone events:
  - Incoming call, outgoing call, call established, call terminated
  - Transfer, hold, unhold, mute, unmute
  - DND on/off, forward on/off
  - Registration success/failure
  - Phone boot, IP change
  - Idle, busy, open DND, close DND
- Variable substitution in URLs ($local, $remote, $display_local, etc.)
- Active URI: incoming HTTP commands to control phone (make call, answer, hangup, etc.)

### Multicast Paging
- Up to 31 multicast paging/listening channels
- Configurable multicast address:port per channel
- Priority levels for paging groups
- Paging barge-in based on priority
- Codec selection for multicast (G.711 typically)

### Network Features
- Dual Gigabit Ethernet ports (LAN + PC pass-through)
- VLAN tagging (802.1Q) for voice and data VLANs separately
- LLDP-MED for auto VLAN discovery
- CDP (Cisco Discovery Protocol) for auto VLAN
- 802.1x authentication (EAP-MD5, EAP-TLS, PEAP-MSCHAPv2)
- PoE (802.3af) — varies by model, T54W is PoE class 2
- IPv4 and IPv6 dual-stack
- DHCP / static IP / PPPoE
- OpenVPN client (on newer firmware)
- LLDP-MED for location identification (E911)

### Diagnostic Tools
- Syslog (UDP) with configurable log levels (0-6)
- SIP log / SIP trace via web UI
- Packet capture (pcap) from web interface
- System log export
- Web UI status pages (network, SIP, call stats)
- Configuration export/import
- LCD diagnostics (key test, LCD test)
- RTP statistics display (jitter, packet loss, MOS)

### Firmware Management
- Manual upload via web UI
- Auto firmware update via provisioning server (TFTP/HTTP/HTTPS)
- Scheduled firmware check
- RPS-triggered firmware updates
- Firmware version check/comparison logic

### Additional Features
- XML browser (for custom phone apps/menus)
- Built-in WiFi (T54W, T53W — 802.11 a/b/g/n/ac dual-band)
- Bluetooth 4.2 (T54W, T53W — headset pairing, contact sync)
- USB port: USB headset, USB recording (call recording to USB stick on some models)
- Color LCD (T54W, T48U) / grayscale (T46U, T43U, T42U)
- HD Voice on handset, speakerphone, headset
- EHS (Electronic Hook Switch) support for wireless headsets
- Hearing Aid Compatible (HAC)
- SIP session timers (RFC 4028)
- SIP REFER (RFC 3515) with Replaces (RFC 3891)
- SIP UPDATE (RFC 3311)
- 100rel / PRACK (RFC 3262)
- SIP outbound (RFC 5626) on newer firmware
- Dialing plan / digit map support
- Emergency call routing (E911 / location headers)
- Multi-language UI (configurable)
- Screensaver / wallpaper customization
- Programmable softkeys (context-sensitive)

---

## 2. Poly (Polycom) VVX Series (VVX 150, 250, 350, 450)

### SIP Registration & Accounts
- Up to 34 SIP registrations on VVX 450; 8 on VVX 350; 8 on VVX 250; 4 on VVX 150
- Per-registration server configuration (primary/secondary/tertiary SIP servers)
- DNS SRV / NAPTR / A record resolution
- Server redundancy with configurable failover and failback
- Registration expiry timer per line
- Simultaneous registration on all lines

### Call Handling
- Hold / Resume
- Music on Hold (local sine tone or server-side)
- Blind transfer (SIP REFER)
- Attended/consultative transfer (REFER with Replaces)
- Semi-attended transfer (transfer during ringback)
- Local 3-way audio conference (built-in bridge)
- Network conference (server-managed via REFER)
- Call park / retrieve (via feature codes or softkeys)
- Directed call pickup / group call pickup
- Last call return
- Call recording trigger (server-controlled)
- Anonymous call / caller ID blocking

### Line Keys / BLF
- Programmable line keys: VVX 450 = 12 line keys (4 pages = 48 total), VVX 350 = 8 line keys (4 pages = 32 total)
- BLF via SIP SUBSCRIBE to dialog event package (RFC 4235)
- BLF with pickup
- Busy Lamp Field list (server-managed BLF lists — SUBSCRIBE to resource list)
- Expansion modules: VVX EM50 (color LCD, 30 keys × 3 pages) — up to 3 modules on VVX 450
- Key types: Line, Speed Dial, BLF, Presence, URL, XML

### Shared Line Appearance (SLA)
- Shared Call Appearance (SCA / BLA)
- Private hold on shared lines
- Barge-in on held shared line calls
- Visual indicators for shared line states (idle, seized, active, held, ringing)

### Presence
- SIP SUBSCRIBE/NOTIFY presence (pidf+xml)
- Buddy watch (presence monitoring list)
- Rich presence with status text
- Microsoft Teams/SfB presence integration (on Edge-E models)

### Intercom / Paging
- Intercom with auto-answer (Alert-Info: Intercom or Auto Answer)
- Push-to-talk intercom mode
- Group paging (via multicast or server)
- Auto-answer configurable via SIP headers: `Alert-Info`, `Call-Info: answer-after=0`

### Call Forwarding
- Unconditional (always)
- On Busy
- On No Answer (ring timer configurable)
- Per-line forwarding rules
- Server-side forwarding integration

### DND
- DND per line or global
- Feature key sync (server-side DND state sync via SUBSCRIBE)
- DND exception list (VIP callers)

### Call Waiting
- Call waiting tone on/off
- Visual + audio indication
- Configurable per line

### Auto-Answer
- Auto-answer via SIP headers: `Alert-Info: Auto Answer`, `Alert-Info: info=alert-autoanswer`, `Call-Info: answer-after=0`
- Auto-answer with mute option
- Auto-answer per-line configurable
- Auto-answer delay

### DTMF
- RFC 2833 (telephone-event)
- SIP INFO (application/dtmf-relay, application/dtmf)
- In-band audio
- Configurable per registration
- DTMF payload type configurable

### Codec Support
- G.711 µ-law / A-law
- G.722 (HD Voice)
- G.729A/B (with license on older models; included on VVX)
- G.726 (all rates: 16/24/32/40)
- iLBC
- Opus (VVX firmware 6.x+)
- Codec negotiation priority configurable
- VAD / CNG
- Acoustic Echo Cancellation
- Full-duplex speakerphone

### Transport & Security
- UDP, TCP, TLS (TLSv1.0/1.1/1.2)
- SRTP (SDES — RFC 4568)
- SRTP modes: optional, required, disabled
- HTTPS for provisioning and web config
- 802.1x (EAP-TLS, EAP-PEAP, EAP-MD5, EAP-FAST)
- Certificate management (import CA, client certs)
- Encrypted configuration files
- Device lock / admin password

### NAT Traversal
- STUN client (RFC 5389)
- Static NAT / fixed IP mapping
- Outbound proxy
- rport (RFC 3581)
- Keep-alive (SIP OPTIONS or CRLF)
- Symmetric RTP
- ICE-Lite (limited models/firmware)

### Provisioning
- Auto-provisioning via TFTP, HTTP, HTTPS, FTP, FTPS
- Poly Zero Touch Provisioning (ZTP) via Poly Lens cloud
- DHCP options 66/160 for boot server
- DHCP option 43 (vendor-specific)
- Provisioning via .cfg XML files (per-phone MAC-based or sip.cfg/phone.cfg common)
- Boot, periodic, and reboot provisioning
- Polycom Provisioning Server (PDMS-SP / RealPresence Resource Manager)

### MWI
- SIP SUBSCRIBE/NOTIFY (RFC 3842)
- LED indicator (flashing)
- On-screen voicemail icon and count
- Stutter dial tone
- One-touch voicemail access
- Per-line voicemail number

### Call History
- Local call log: placed, received, missed
- Typically 100+ entries
- Call duration display
- Server-side call log integration (BroadSoft, etc.)

### Phonebook / Directory
- Local contact directory: up to 500+ contacts (VVX 450)
- Corporate directory (LDAP, LDAPS)
- BroadSoft directory (XSI)
- Remote contact directory (XML over HTTP/HTTPS)
- Favorites list
- Speed dial
- Contact search across all directories
- Import/export contacts

### QoS
- DSCP/DiffServ configurable for SIP signaling and RTP independently
- Default: RTP=EF (46), SIP=AF31 (26)
- 802.1p Layer 2 priority tagging
- LLDP-MED for automatic QoS policy from switch

### Hotdesking
- Hotdesking / flexible seating (clears user config, prompts new login)
- BroadSoft Flexible Seating guest/host
- Server-side hotdesking support

### Action URLs
- Web/HTTP notifications (limited compared to Yealink/Snom)
- Polling-based phone state API (UCS REST API on some firmware)
- Push-action URI (limited event set)

### Multicast Paging
- Up to 25 multicast paging groups
- Configurable multicast address:port per group
- Priority-based paging (higher priority interrupts lower)
- Emergency page priority
- Codec configurable per group

### Network Features
- Dual Gigabit Ethernet (VVX 250/350/450), dual 10/100 (VVX 150)
- VLAN tagging (802.1Q) — separate voice and data VLANs
- LLDP-MED for auto VLAN discovery
- CDP for auto VLAN
- 802.1x port-based authentication
- PoE (802.3af / 802.3at on some models)
- IPv4 / IPv6 dual-stack
- DHCP / Static IP

### Diagnostic Tools
- Syslog (UDP, configurable facility and level)
- SIP signaling trace / protocol logging
- Upload logs to server
- Web UI status pages (call stats, network, SIP status)
- Configuration report
- Packet capture (limited — not all models)
- MOS / call quality statistics per call
- Problem report generation

### Firmware Management
- Manual upload via web UI
- Auto firmware update via provisioning server
- Poly Lens cloud-based firmware management
- Scheduled update windows
- Firmware downgrade protection (configurable)

### Additional Features
- XHTML/XML microbrowser for custom apps
- Built-in WiFi (VVX 250/350/450 have optional WiFi dongle or built-in on some SKUs)
- Bluetooth (VVX 350/450 via USB dongle)
- USB ports for headsets and accessories
- Color LCD (VVX 250/350/450)
- HD Voice (wideband audio on handset, headset, speaker)
- EHS support (Plantronics/Poly, Jabra)
- SIP session timers (RFC 4028)
- SIP REFER (RFC 3515)
- SIP Replaces (RFC 3891)
- SIP UPDATE (RFC 3311)
- 100rel / PRACK (RFC 3262)
- Digit map / dialing rules
- Emergency call routing
- Multi-language support
- Customizable backgrounds/themes
- Accessibility features (visual alerts, HAC)

---

## 3. Cisco IP Phone 7800 / 8800 Series — Multiplatform (MPP) Firmware

(7821, 7841, 7861, 8811, 8841, 8845, 8851, 8861, 8865)

### SIP Registration & Accounts
- 7821: 2 lines; 7841: 4 lines; 7861: 16 lines
- 8811: 1 line; 8841: 4 lines; 8845: 4 lines; 8851: 10 lines; 8861: 16 lines; 8865: 16 lines
- Per-line SIP server config (primary, secondary, tertiary)
- DNS SRV / NAPTR / A record
- Registration expiry per line
- Failover/failback between servers
- Simultaneous multi-line registration

### Call Handling
- Hold / Resume
- Music on Hold (server-side or local tone)
- Blind transfer (REFER)
- Attended/consultative transfer (REFER with Replaces)
- Semi-attended transfer
- Local 3-way conference (built-in bridge)
- N-way conference (server-side)
- Call park / retrieve (via softkey or star code)
- Directed call pickup / group call pickup
- Call back on busy
- Call recording (server-triggered)
- Anonymous call / caller ID blocking (Privacy header)
- Plus Dialing (E.164 with ‘+’ prefix)

### Line Keys / BLF
- Programmable line keys (model-dependent, see line counts above)
- BLF via SUBSCRIBE to dialog event package
- Speed dial + BLF combined
- BLF directed call pickup
- KEM (Key Expansion Module): BEKEM for 8800 series (36 keys × 2 pages = 72 per module, up to 3 modules = 216 keys)
- 7800 series: no expansion module support
- Key functions: Line, Speed Dial, BLF, Call Park, Intercom

### Shared Line Appearance
- Shared line (SCA / BLA)
- Private hold on shared lines
- Visual indicators for shared line state
- Barge-in support (configurable)

### Presence
- BLF-based presence (dialog state)
- SUBSCRIBE/NOTIFY for line state
- XMPP presence (when integrated with Webex/BroadSoft)

### Intercom / Paging
- Intercom with auto-answer
- Supported headers: `Call-Info: answer-after=0`, `Alert-Info`
- Whisper intercom (one-way or two-way)
- Group paging

### Call Forwarding
- Call Forward All
- Call Forward Busy
- Call Forward No Answer (configurable timer)
- Per-line forwarding
- Feature key sync for forwarding state

### DND
- DND on/off (per line or global)
- Feature key synchronization with server
- DND with star code

### Call Waiting
- Call waiting on/off per line
- Audio + visual notification

### Auto-Answer
- Auto-answer via SIP headers: `Call-Info: answer-after=0`, `Alert-Info`
- Auto-answer with mute
- Auto-answer per-line

### DTMF
- RFC 2833 (telephone-event)
- SIP INFO (application/dtmf-relay)
- In-band (G.711)
- AVT (named telephone event)
- Configurable per line

### Codec Support
- G.711 µ-law / A-law
- G.722 (wideband)
- G.729a/b
- G.726
- iLBC (20ms / 30ms)
- Opus (8800 series on newer firmware)
- Codec priority list configurable
- VAD / CNG
- Comfort noise generation
- Jitter buffer (adaptive)

### Transport & Security
- UDP, TCP, TLS (TLSv1.1/1.2)
- SRTP (SDES — RFC 4568)
- Cisco SRST (Survivable Remote Site Telephony) — MPP limited
- HTTPS for provisioning and web admin
- 802.1x (EAP-TLS, EAP-FAST, EAP-PEAP, EAP-MD5)
- MIC (Manufacturing Installed Certificate) for secure provisioning
- SCEP certificate enrollment (on some firmware)
- Signed firmware images

### NAT Traversal
- STUN (RFC 5389)
- Outbound proxy
- rport (RFC 3581)
- NAT keep-alive (UDP/SIP)
- Symmetric RTP
- Static NAT mapping
- ICE (limited support on newer firmware)

### Provisioning
- Auto-provisioning via TFTP, HTTP, HTTPS
- Cisco EDOS (Enhanced Device Onboarding Service) — zero-touch cloud redirect
- DHCP option 66/150/159/160
- TR-069 (CWMP)
- XML-based config profiles ($MA.cfg, $PSN.xml)
- Per-phone (MAC-based) and common config files
- Resync on timer / at specific time

### MWI
- SUBSCRIBE/NOTIFY (RFC 3842)
- LED indicator (red light on handset)
- On-screen icon + count
- Stutter dial tone
- Voicemail key / speed dial

### Call History
- Local call log: all, placed, received, missed
- 150 entries per category
- Call duration
- BroadSoft call log (XSI)

### Phonebook / Directory
- Local directory (up to 200 entries)
- LDAP directory (LDAP/LDAPS)
- XML directory service (HTTP/HTTPS)
- BroadSoft XSI directory
- Personal directory (on server)
- Speed dial list
- Search across directories

### QoS
- DSCP for SIP and RTP configurable independently
- Default: RTP=EF(46), SIP=AF31(26)
- 802.1p priority
- LLDP-MED for auto QoS
- CDP for auto QoS

### Hotdesking
- Extension Mobility (login/logout)
- Guest login
- BroadSoft Flexible Seating

### Action URLs
- Limited action URL / webhook support in MPP mode
- XML service push
- Phone can receive XML display objects via HTTP push

### Multicast Paging
- Up to 10+ multicast paging groups
- Configurable multicast address per group
- Priority-based (higher priority interrupts active page or call)
- Auto-answer for paging

### Network Features
- Dual Gigabit Ethernet (most 8800 models), 10/100 on 7800
- VLAN (802.1Q) — voice VLAN + data VLAN
- LLDP-MED
- CDP (native Cisco support)
- 802.1x
- PoE (802.3af, 802.3at on some 8800)
- IPv4 / IPv6 dual-stack
- DHCP / Static

### Diagnostic Tools
- Syslog (UDP, configurable levels)
- SIP message logging (debug level)
- PRT (Problem Report Tool) — generates log bundle, uploads to server
- Web UI status and statistics
- Call statistics display (per-call jitter, loss, MOS)
- Packet capture (via SPAN on switch, not natively on phone in MPP)
- Configuration report

### Firmware Management
- Manual upload via web UI
- Auto firmware update via provisioning (TFTP/HTTP/HTTPS)
- Cisco firmware.webex.com / cloud-managed updates
- Firmware load rules in provisioning config
- Signed firmware enforcement

### Additional Features
- WiFi: 8845/8865 (802.11 a/b/g/n/ac)
- Bluetooth: 8845/8865 (BT for headset + mobile pairing)
- USB: 8851/8861/8865 (USB headset, charging, KEM)
- Video: 8845/8865 (720p HD video calling)
- Color LCD (8800 series), grayscale (7800 series)
- HD Voice wideband
- EHS for wireless headsets
- SIP session timers (RFC 4028)
- SIP REFER (RFC 3515) + Replaces
- SIP UPDATE (RFC 3311)
- 100rel / PRACK (RFC 3262)
- XML services (push XML display)
- Digit map / dial plan
- Emergency call routing (E911 with LLDP location)
- Multi-language
- Cisco Webex integration (on MPP firmware)
- Noise removal (8800 series)
- Adjustable ring tones / custom ring tones

---

## 4. Grandstream GRP2600 Series & GXP Series

(GRP2612, GRP2613, GRP2614, GRP2615, GRP2616, GXP2130, GXP2135, GXP2160, GXP2170)

### SIP Registration & Accounts
- GRP2612/2613: 4 SIP accounts; GRP2614: 4 accounts; GRP2615: 10 accounts; GRP2616: 6 accounts
- GXP2130: 3 accounts; GXP2135: 8 accounts; GXP2160: 6 accounts; GXP2170: 12 accounts
- Per-account SIP server configuration (primary + failover)
- DNS SRV / NAPTR / A record lookup
- Registration expiry configurable
- Failover between primary/secondary/tertiary SIP servers
- Concurrent registration on all accounts

### Call Handling
- Hold / Resume
- Blind transfer (REFER)
- Attended transfer (REFER with Replaces)
- Semi-attended transfer
- Local 5-way conference (GRP2614/2615/2616 and GXP2135/2160/2170)
- Local 3-way conference (other models)
- Call park / retrieve (via feature code)
- Directed call pickup / group call pickup
- Call completion (CCBS)
- Anonymous call (Privacy header)
- Anonymous call rejection
- Call recording (server-triggered)
- Click-to-dial

### Line Keys / BLF
- GRP2612: 4 dual-color line keys (up to 2 pages on some models)
- GRP2614: 4 line keys + 10 BLF keys on secondary LCD
- GRP2615: 10 line keys + 40 virtual multi-purpose keys (VPK, multiple pages)
- GRP2616: 6 line keys + 24 BLF keys on secondary display
- GXP2170: 12 line keys + 48 virtual multi-purpose keys
- BLF via SUBSCRIBE to dialog event package (RFC 4235)
- BLF-to-pickup
- Eventlist BLF (resource list SUBSCRIBE)
- Key types: Line, Speed Dial, BLF, Presence, Call Park, Intercom, DTMF, Voicemail, Call Transfer, Multicast Paging, Monitored Call Park
- Expansion module: GBX20 (20 keys × 2 pages per module, up to 4 modules)
- Virtual Multi-Purpose Keys (VPK) — software-based BLF pages (up to 20+ pages)

### Shared Line Appearance
- Shared Call Appearance (SCA/BLA)
- BroadSoft SCA support
- Private hold per shared line
- Visual indicators (idle, seized, active, held, ringing)

### Presence
- SIP SUBSCRIBE/NOTIFY (presence event package)
- BLF presence
- Busy indication on speed dial keys

### Intercom / Paging
- Intercom with auto-answer
- Alert-Info / Call-Info header based auto-answer
- Full-duplex intercom
- Group paging

### Call Forwarding
- Unconditional
- On Busy
- On No Answer (configurable timer)
- Time-based forwarding rules
- Per-account forwarding

### DND
- DND on/off per account or global
- DND timer (auto-disable after duration)
- Feature key sync

### Call Waiting
- Call waiting on/off
- Call waiting tone configurable

### Auto-Answer
- Auto-answer via SIP headers: `Call-Info: answer-after=0`, `Alert-Info`
- Auto-answer per account
- Auto-answer with mute

### DTMF
- RFC 2833 (telephone-event)
- SIP INFO (application/dtmf-relay)
- In-band audio
- Configurable per account
- Payload type configurable

### Codec Support
- G.711 µ-law / A-law
- G.722 (wideband)
- G.729A/B
- G.726-16/24/32/40
- iLBC (20ms / 30ms)
- Opus (GRP series)
- G.723.1 (GXP series)
- Codec priority configurable per account
- VAD / CNG / AGC
- Jitter buffer (fixed/adaptive)
- AEC / noise reduction

### Transport & Security
- UDP, TCP, TLS (TLSv1.0/1.1/1.2)
- SRTP (SDES — RFC 4568)
- SRTP modes: enabled/disabled/optional per account
- HTTPS for web management and provisioning
- 802.1x (EAP-MD5, EAP-TLS, EAP-PEAP)
- Certificate import (trusted CA, device cert)
- Encrypted configuration files (AES-256)

### NAT Traversal
- STUN server (RFC 5389)
- Outbound proxy
- rport (RFC 3581)
- NAT keep-alive (UDP/SIP)
- Symmetric RTP
- Static NAT
- ICE (on GRP series newer firmware)

### Provisioning
- Auto-provisioning via TFTP, HTTP, HTTPS, FTP
- Grandstream GDMS (Grandstream Device Management System) — cloud zero-touch
- DHCP option 66/43/120/160
- TR-069 (CWMP)
- Plug-and-Play via SIP multicast (GRP series)
- XML/JSON config files (per-phone MAC-based + global)
- AES-encrypted provisioning
- Scheduled auto-provisioning

### MWI
- SUBSCRIBE/NOTIFY (RFC 3842)
- LED indicator
- On-screen icon + voicemail count
- Stutter dial tone
- Voicemail access key per account

### Call History
- Local call log: placed, received, missed
- Up to 2000 entries (GRP series), 500 (GXP)
- Call duration and timestamp
- Export call history

### Phonebook / Directory
- Local phonebook: up to 2000 contacts (GRP), 1000 (GXP2170)
- LDAP directory (LDAP/LDAPS)
- Remote XML phonebook (HTTP/HTTPS) — multiple URLs
- BroadSoft XSI directory
- Favorites
- Blacklist
- Import/export via XML/CSV/vCard
- Phonebook search across all sources

### QoS
- DSCP for SIP signaling and RTP independently
- 802.1p Layer 2 priority
- LLDP for QoS policy discovery
- Default: RTP=EF(46), SIP=AF31(26)

### Hotdesking
- Hotdesking feature (clear + re-register with new credentials)
- SIP SUBSCRIBE-based hotdesking
- ACD (Automatic Call Distribution) support

### Action URLs
- HTTP/HTTPS action URLs on phone events:
  - Incoming call, outgoing call, call connected, call disconnected
  - On hold, off hold, DND, forwarding, transfer
  - Registration, boot, etc.
- Variable substitution
- GXP Toolkit for custom app integration

### Multicast Paging
- Up to 10 multicast paging listening addresses
- Priority levels per group (priority page interrupts lower)
- Multicast paging send
- Codec configurable for multicast

### Network Features
- Dual Gigabit Ethernet (GRP and GXP2130+)
- VLAN tagging (802.1Q) — separate voice/data
- LLDP-MED for VLAN discovery
- CDP support
- 802.1x authentication
- PoE (802.3af)
- IPv4 / IPv6
- DHCP / Static / PPPoE

### Diagnostic Tools
- Syslog (UDP, configurable level)
- SIP log / debug logging
- Packet capture (pcap) via web UI (GRP series, newer GXP)
- System log download
- Web UI status (network, SIP registrations, call stats)
- LCD diagnostic screen
- RTP statistics per call

### Firmware Management
- Manual upload via web UI
- Auto firmware update via provisioning server
- GDMS cloud firmware management
- Firmware upgrade/downgrade control
- Scheduled firmware checks

### Additional Features
- WiFi (GRP2612W, GRP2614, GRP2615, GRP2616 — 802.11 a/b/g/n/ac)
- Bluetooth (GRP2614, GRP2615, GRP2616 — BT headset pairing)
- USB port (GRP2614/2615/2616 — headset, recording)
- Color LCD (GRP2614/2615/2616, GXP2160/2170), grayscale (others)
- HD Voice
- EHS for wireless headsets (Plantronics)
- SIP session timers (RFC 4028)
- SIP REFER (RFC 3515) + Replaces (RFC 3891)
- SIP UPDATE (RFC 3311)
- 100rel / PRACK (RFC 3262)
- XML custom screen applications
- Digit map / dial plan
- Screensaver / wallpaper
- Multi-language UI
- Emergency call (E911 location via LLDP)
- Noise shield / noise reduction
- Swappable faceplate (GRP2612/2613)

---

## 5. Snom D7xx Series (D713, D715, D717, D735, D765, D785)

### SIP Registration & Accounts
- D713: 4 SIP accounts; D715: 4 accounts; D717: 4 accounts
- D735: 12 accounts; D765: 12 accounts; D785: 12 accounts
- Per-identity (account) SIP server configuration
- DNS SRV / NAPTR / A record
- Registration expiry per identity
- Failover SIP servers (backup registrar)
- Simultaneous registration on all identities

### Call Handling
- Hold / Resume (RFC 3264)
- Music on Hold (local file or server)
- Blind transfer (REFER)
- Attended transfer (REFER with Replaces)
- Semi-attended transfer
- Local 3-way conference (built-in mixer)
- Local 5-way conference (D785)
- Conference via server (REFER-based)
- Call park / retrieve (via star codes or BLF)
- Directed call pickup / group call pickup
- Call completion on busy (CCBS) / no reply (CCNR)
- Anonymous call (caller ID suppression — Privacy: id)
- Anonymous call rejection
- Call deflection (302 redirect)
- Call recording (server-triggered or on-phone recording on some models)

### Line Keys / BLF
- D713: 5 programmable function keys; D715: 5 keys; D717: 18 keys (with secondary display)
- D735: 32 keys (with color secondary display); D765: 16 keys; D785: 24 keys (with touch secondary display)
- Multi-page function keys (up to 42+ on D785 with pages)
- BLF via SUBSCRIBE to dialog event package (RFC 4235)
- Eventlist BLF (resource list SUBSCRIBE — RFC 4662)
- BLF with pickup
- Expansion modules: D7 (18 keys per module, monochrome), D7C (18 keys, color)
  - Up to 3 modules per phone = 54 additional keys
- Key types: Line, BLF, Speed Dial, Intercom, Park Orbit, Park+Orbit, DTMF, Action URL, Extension, Presence, Button, Custom

### Shared Line Appearance
- SCA (Shared Call Appearance)
- BLA (Bridged Line Appearance)
- Private hold on shared lines
- Visual indicators for line state
- Full SCA support per Broadcom/BroadSoft SCA draft

### Presence
- SIP SUBSCRIBE/NOTIFY (presence event package, RFC 3856)
- Rich presence (pidf+xml)
- BLF-based presence monitoring
- Presence status on function keys

### Intercom / Paging
- Intercom with auto-answer
- Supported headers: `Alert-Info`, `Call-Info: answer-after=0`, `Answer-Mode: Auto` (RFC 5373)
- Snom is one of few phones supporting RFC 5373 Answer-Mode header natively
- Full-duplex intercom
- Push-to-talk

### Call Forwarding
- Unconditional (always)
- On Busy
- On No Answer (configurable timeout)
- Per-identity forwarding
- On-phone forwarding UI

### DND
- DND per identity or global
- DND with ringer off but visual notification
- Feature key sync with server

### Call Waiting
- Call waiting on/off
- Call waiting tone
- Visual indication

### Auto-Answer
- Auto-answer via SIP headers:
  - `Alert-Info: info=alert-autoanswer`
  - `Alert-Info: <http://www.notused.com>;info=alert-autoanswer`
  - `Call-Info: ;answer-after=0`
  - `Answer-Mode: Auto` (RFC 5373 — Snom specialty)
- Auto-answer delay configurable
- Auto-answer with mute
- Per-identity auto-answer setting

### DTMF
- RFC 2833 (telephone-event)
- SIP INFO (application/dtmf-relay)
- SIP INFO (application/dtmf)
- In-band audio
- Configurable per identity
- DTMF payload type configurable

### Codec Support
- G.711 µ-law / A-law
- G.722 (wideband)
- G.729A/B
- G.726 (all sub-rates)
- iLBC
- Opus (on newer firmware, D7xx series)
- L16 (16-bit linear PCM — on some firmware)
- Codec priority list per identity
- VAD / CNG
- AEC / noise reduction
- AGC (Automatic Gain Control)

### Transport & Security
- UDP, TCP, TLS (TLSv1.0/1.1/1.2/1.3)
- SRTP (SDES — RFC 4568)
- SRTP modes: on/off/optional per identity
- DTLS-SRTP (on newer firmware — Snom is one of few desk phones with DTLS support)
- HTTPS for web and provisioning
- 802.1x (EAP-TLS, EAP-PEAP, EAP-MD5)
- Client certificate / CA certificate management
- SRTP key renegotiation
- SIPS URI support

### NAT Traversal
- STUN (RFC 5389)
- ICE (RFC 8445 — on newer firmware)
- Outbound proxy
- rport (RFC 3581)
- Symmetric RTP
- NAT keep-alive (SIP/UDP)
- Static NAT mapping
- SIP outbound (RFC 5626)

### Provisioning
- Auto-provisioning via TFTP, HTTP, HTTPS, FTP
- Snom SRAPS (Snom Redirect and Provisioning Service) — cloud zero-touch
- DHCP option 66/67/43
- Provisioning via XML config files (per-phone MAC-based, model-based, common)
- PnP via SIP multicast SUBSCRIBE
- TR-069 (CWMP) — on some firmware
- Scheduled re-provisioning
- Setting protection (lock specific settings from being overwritten)
- Encrypted provisioning (HTTPS with client certs)

### MWI
- SUBSCRIBE/NOTIFY (RFC 3842)
- LED indicator (dedicated MWI LED)
- On-screen icon + voicemail count (messages-waiting summary)
- Stutter dial tone
- Per-identity voicemail number

### Call History
- Local call log: dialed, received, missed
- Typically 100 entries per category
- Server-side call log integration

### Phonebook / Directory
- Local directory: up to 1000 contacts
- LDAP/LDAPS directory
- Remote XML phonebook (HTTP/HTTPS)
- XCAP (XML Configuration Access Protocol) for network-stored contacts
- Favorites
- Blacklist (call blocking)
- Speed dial
- Contact search across all sources
- Import/export contacts

### QoS
- DSCP/DiffServ configurable for SIP and RTP independently
- 802.1p priority tagging
- LLDP-MED for auto QoS and VLAN
- Default: RTP=EF(46), SIP=AF31(26)

### Hotdesking
- Hotdesking (clear registration, prompt re-login)
- Per-identity hotdesking
- Server-controlled hotdesking

### Action URLs (Snom Specialty — Industry-Leading)
- Snom has the most extensive Action URL support of any SIP phone:
- HTTP/HTTPS GET/POST on dozens of events:
  - **Call events:** incoming, outgoing, connected, disconnected, missed, offhook, onhook
  - **Call control:** hold, unhold, transfer, conference
  - **Feature events:** DND on/off, forward on/off, mute/unmute
  - **Registration:** register ok, register failed, unregister
  - **System:** setup complete (boot), IP change, reboot
  - **Key events:** function key press, DTMF digit press
  - **Presence:** presence change
  - **Sensor:** motion sensor (D785)
- Variable substitution: $local, $remote, $active_url, $csta_id, $display_remote, etc.
- **Incoming Action URL:** Phone can receive HTTP commands to control behavior (make call, answer, transfer, hangup, set LED, display message, etc.)
- Can render returned XML/HTML content on phone display
- CSTA (Computer Supported Telecommunications Applications) support
- Snom XML minibrowser with event-driven navigation

### Multicast Paging
- Up to 10+ multicast paging addresses
- Priority-based paging (interrupts calls if higher priority)
- Multicast send and receive
- Codec configurable

### Network Features
- Dual Gigabit Ethernet (D735, D765, D785, D717), 10/100 (D713, D715)
- VLAN tagging (802.1Q) — separate voice and data VLANs
- LLDP-MED for auto VLAN and QoS
- CDP support
- 802.1x (EAP-TLS, PEAP, MD5)
- PoE (802.3af)
- IPv4 / IPv6 dual-stack
- DHCP / Static IP / PPPoE
- VPN client (OpenVPN on some firmware versions)

### Diagnostic Tools
- Syslog (UDP, TCP) with 10 log levels
- SIP trace (full SIP message logging via syslog)
- Packet capture (pcap) directly from phone web UI — Snom does this very well
- On-phone pcap capture downloadable from web interface
- System information page (web UI)
- SIP trace viewer in web UI
- Call quality statistics (MOS, jitter, packet loss per call)
- Configuration backup/restore
- Debug logging categories (SIP, RTP, HTTP, SRTP, etc.)

### Firmware Management
- Manual upload via web UI
- Auto firmware update via provisioning server
- SRAPS cloud firmware management
- Firmware scheduled update
- Beta/stable firmware tracks

### Additional Features
- XML minibrowser (Snom specialty — render custom XML UI on phone screen via HTTP)
- Bluetooth (D735, D785 — BT headset + handset + device pairing)
- WiFi (D735 via optional USB dongle)
- USB port: USB headset, USB storage
- Color LCD (D735, D785, D717), grayscale (D713, D715)
- Secondary display (D717: small LCD for BLF; D785: touch BLF display)
- HD Voice (wideband handset, speaker, headset)
- EHS (Electronic Hook Switch) for wireless headsets
- SIP session timers (RFC 4028)
- SIP REFER (RFC 3515) + Replaces (RFC 3891)
- SIP UPDATE (RFC 3311)
- 100rel / PRACK (RFC 3262)
- SIP outbound (RFC 5626) — Snom is notable for RFC 5626 support
- GRUU (Globally Routable UA URI) support
- SIP Path header support
- CSTA (uaCSTA over SIP) for CTI integration
- Digit map / dial plan with regex
- Emergency call (E911, location headers)
- Multi-language (20+ languages)
- Customizable idle screen / screensaver
- Sensor-based features (D785 proximity sensor for auto wake)
- Per-identity ringtones and settings
- Customizable softkeys per call state

---

## 6. Cross-Vendor Feature Comparison Summary

### Universal Features (ALL vendors support)
- Multiple SIP account registration with failover
- DNS SRV/A record server resolution
- Hold/Resume, Blind Transfer (REFER), Attended Transfer (REFER+Replaces)
- Local 3-way conference
- BLF via dialog event package (RFC 4235)
- Shared Call Appearance / BLA
- Call forwarding (unconditional/busy/no-answer)
- DND
- Call waiting
- MWI (RFC 3842)
- DTMF: RFC 2833, SIP INFO, In-band
- Codecs: G.711a/u, G.722, G.729, G.726
- Transport: UDP, TCP, TLS
- SRTP (SDES/RFC 4568)
- STUN, rport, outbound proxy, symmetric RTP
- Auto-provisioning (TFTP/HTTP/HTTPS)
- DHCP option 66 for boot server
- Cloud zero-touch provisioning (vendor-specific: RPS/ZTP/EDOS/GDMS/SRAPS)
- Local call history (placed/received/missed)
- Local contacts + LDAP + remote XML phonebook
- DSCP/QoS markings
- VLAN tagging (802.1Q)
- LLDP-MED, CDP
- 802.1x authentication
- PoE (802.3af)
- Syslog
- SIP session timers (RFC 4028)
- SIP REFER, Replaces, UPDATE, 100rel/PRACK
- Digit map / dial plan
- Auto-answer via SIP headers (Call-Info: answer-after=0)
- Multicast paging
- Expansion modules (except Cisco 7800)

### Differentiating Features by Vendor

| Feature | Yealink | Polycom | Cisco MPP | Grandstream | Snom |
|---|---|---|---|---|---|
| Max SIP accounts | 16 | 34 | 16 | 12 | 12 |
| Local conference | 5-way | 3-way | 3-way | 5-way | 5-way |
| Opus codec | Yes (newer FW) | Yes (6.x+) | Yes (8800 newer) | Yes (GRP) | Yes (newer FW) |
| iLBC | Yes | Yes | Yes | Yes | Yes |
| DTLS-SRTP | No | No | No | No | Yes (Snom!) |
| ICE | Limited | Limited | Limited | Limited | Yes (newer FW) |
| RFC 5626 Outbound | Newer FW | No | No | No | Yes |
| RFC 5373 Answer-Mode | No | No | No | No | Yes (Snom!) |
| Action URLs (extensive) | Good | Limited | Limited | Good | Best in class |
| Active URI (inbound HTTP) | Yes | Limited | XML push | Yes | Yes (best) |
| XML browser/apps | Yes | XHTML micro | XML services | Yes | XML minibrowser |
| Packet capture on phone | Yes | Limited | No (MPP) | Yes (GRP) | Yes (best) |
| CSTA | No | No | No | No | Yes |
| GRUU | No | No | No | No | Yes |
| TR-069 | Yes | No | Yes | Yes | Limited |
| Video calling | No | No | 8845/8865 | No | No |
| OpenVPN | Yes (newer) | No | No | No | Yes (limited) |
| Multicast page groups | 31 | 25 | 10+ | 10 | 10+ |
| BLF key capacity (w/EXP) | 180+ | 270+ | 216 (8800) | 160+ | 96+ |
| WiFi built-in | T54W/T53W | Select SKUs | 8845/8865 | GRP W models | Via dongle |
| Bluetooth | T54W/T53W | Via dongle | 8845/8865 | GRP2614+ | D735/D785 |

### Key SIP RFCs Supported Across All
- RFC 3261 — SIP core
- RFC 3262 — 100rel / PRACK
- RFC 3263 — SIP server location (DNS)
- RFC 3264 — SDP offer/answer
- RFC 3265 — SIP SUBSCRIBE/NOTIFY
- RFC 3311 — UPDATE
- RFC 3323 — Privacy
- RFC 3325 — P-Asserted-Identity
- RFC 3515 — REFER
- RFC 3581 — rport
- RFC 3842 — MWI
- RFC 3856 — Presence
- RFC 3891 — Replaces
- RFC 4028 — Session Timers
- RFC 4235 — Dialog Event Package (BLF)
- RFC 4568 — SRTP SDES
- RFC 5389 — STUN

### Snom-Only or Snom-Notable RFCs
- RFC 5373 — Answer-Mode header
- RFC 5626 — SIP Outbound
- RFC 5627 — GRUU
- DTLS-SRTP (RFC 5764)
- uaCSTA over SIP

### Implications for Our SIP Softphone Implementation
To match hardware phone feature parity, our SIP UA should support at minimum:
1. Multi-account SIP registration with failover
2. All transfer types (blind, attended, semi-attended) via REFER+Replaces
3. Local conference mixing (at least 3-way)
4. BLF via SUBSCRIBE to dialog event package
5. SUBSCRIBE/NOTIFY for MWI, presence
6. Hold/resume with proper SDP manipulation
7. DTMF: RFC 2833 + SIP INFO
8. Codecs: G.711, G.722, G.729 at minimum; Opus for modern deployments
9. TLS + SRTP (SDES)
10. Auto-answer via Call-Info/Alert-Info header detection
11. Call forwarding, DND, call waiting
12. NAT traversal: STUN, rport, symmetric RTP, outbound proxy
13. SIP session timers (RFC 4028)
14. 100rel/PRACK (RFC 3262)
15. Call history logging
16. Contact/directory management
