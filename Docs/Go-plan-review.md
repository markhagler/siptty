# Go Plan Review — Gaps, Risks, and Custom Work

Date: 2026-02-09

This review compares `Docs/plans/2026-02-09-go-rewrite-design.md` to
`Docs/DESIGN.md` and calls out areas where the Go plan needs custom code or
re-scoping. The Go plan is directionally aligned with the product goals and UI
layout, but several capabilities in the Python/PJSIP design do not map cleanly
onto sipgo/diago without additional investment.

## Summary

The Go plan should work for an MVP focused on basic registration and single
call control with G.711 media. However, it assumes more SIP trace visibility
and media control than sipgo/diago provide out of the box. Tier 2 and Tier 3
features from `Docs/DESIGN.md` will require significant custom SIP plumbing and
media work.

## Major gaps vs `Docs/DESIGN.md`

1. **Browser audio is missing.**
   - `Docs/DESIGN.md` includes a browser audio mode (WebSocket bridge +
     WebAudio API). The Go plan explicitly drops it.
   - If browser audio is a key differentiator, it must be restored as a planned
     Phase 3 feature with a concrete architecture (RTP <-> WebSocket bridge,
     jitter buffering, device selection, codec constraints).

2. **SIP trace capture may not be sufficient.**
   - The Go plan assumes “sipgo debug logging hooks” can provide full raw SIP
     messages with direction, Call-ID, etc.
   - sipgo logging may not expose structured raw message payloads in the form
     needed for a reliable sngrep-style viewer.
   - The dialog viewer depends on high-fidelity trace data. Without a stable
     hook, this feature becomes fragile or infeasible.

3. **Media control assumptions are optimistic.**
   - The plan lists mute/unmute, DTMF send/receive, WAV playback/recording as
     direct diago calls.
   - diago’s media pipeline is limited; “simple RTP” is fine, but advanced
     media control and recording/playback workflows can require custom RTP
     handling and buffering.

4. **Multi-call and hold/resume are deferred but impact architecture now.**
   - The current UI and engine layout read as single-call focused, while
     multi-call handling is explicitly deferred.
   - If the data model and TUI wiring are not prepared for multiple calls,
     Phase 2 will require rework rather than incremental additions.

5. **Header override parity is not guaranteed.**
   - `Docs/DESIGN.md` expects a robust header override pipeline including
     runtime editing. The Go plan limits overrides to config-level only.
   - If header spoofing and runtime mutation are core “SIP swiss-army-knife”
     goals, this is a meaningful reduction in capability.

## Areas likely to need custom code

1. **Browser audio (desired feature)**
   - Build WebSocket transport for audio frames.
   - Implement jitter buffer and clock sync.
   - Encode/decode to PCM for WebAudio, optionally transcode codecs.
   - Maintain call audio routing between RTP and browser.

2. **SIP trace and dialog viewer plumbing**
   - Integrate at sipgo’s message processing layer, not just log output.
   - Preserve raw SIP messages with direction + timestamps.
   - Extract structured headers reliably (Call-ID, From, To, CSeq, status).

3. **Hold/resume and attended transfer**
   - Implement re-INVITE with SDP direction attributes.
   - Track dialog state transitions and update TUI accordingly.

4. **BLF/Presence/MWI subscriptions**
   - Add SUBSCRIBE/NOTIFY transaction support if diago lacks it.
   - Implement dialog lifecycle, refresh timers, and event parsing.

5. **Conference / multi-party**
   - diago is 2-party oriented; true conferencing requires audio mixing or
     external media bridge.

6. **Codec expansion**
   - G.711 is fine for MVP. Adding Opus/G.722 likely requires CGO or a pure-Go
     codec implementation and SDP negotiation support.

## Recommendations to improve feasibility

1. Add a “trace capture feasibility spike” to validate access to raw SIP
   messages before building the dialog viewer.
2. Make browser audio an explicit Phase 3 goal if it remains important.
3. Acknowledge diago limits and add a “custom media pipeline” placeholder for
   the advanced roadmap.
4. Design the call registry and TUI around multiple simultaneous calls from
   day 1, even if the UI exposes only one in MVP.
5. Add a feature mapping table that enumerates which `Docs/DESIGN.md` features
   are MVP, Phase 2, Phase 3, or “requires custom SIP stack work.”

## Feature mapping (DESIGN.md → Go plan)

Legend for “Go plan status”:
- MVP = in scope for Phase 1 in the Go plan
- P2 = deferred to Phase 2
- P3 = deferred to Phase 3
- Custom = requires substantial custom SIP/media work beyond sipgo/diago
- Dropped = not included in the Go plan

| Feature (DESIGN.md) | Go plan status | Notes |
|---|---|---|
| Register / Unregister | MVP | Matches diago `Register()` + context cancel. |
| Outbound / inbound calls | MVP | Core diago flow. |
| Hangup | MVP | diago `Hangup()`. |
| Call hold / resume | P2 + Custom | Needs re-INVITE with SDP direction. |
| Mute / unmute | MVP (risk) | Depends on diago media control; may need custom RTP. |
| Send DTMF | MVP (risk) | diago `AudioWriterDTMF()`. |
| Receive DTMF | MVP (risk) | diago `AudioReaderDTMF()`. |
| SIP trace display | MVP (risk) | Requires reliable raw SIP hook. |
| SIP dialog viewer | MVP (risk) | Depends on trace fidelity. |
| File audio play/record | MVP (risk) | Diago media pipeline limitations. |
| Null audio mode | MVP | No media reader/writer. |
| BLF subscriptions | P2 + Custom | SUBSCRIBE/NOTIFY support needed. |
| MWI | P2 + Custom | SUBSCRIBE/NOTIFY support needed. |
| Presence pub/sub | P3 + Custom | SUBSCRIBE/PUBLISH handling. |
| Blind transfer | MVP | diago `Refer()`. |
| Attended transfer | P2 + Custom | Needs Replaces handling. |
| 3-way conference | P2 + Custom | Audio mixing/bridge needed. |
| Multiple simultaneous calls | P2 | Requires engine + UI support. |
| Call history (SQLite) | P3 | App-layer only. |
| Runtime header editor | P3 + Custom | Requires per-request header mutation. |
| Config header overrides | MVP | Config-level only. |
| User-Agent spoofing | MVP | Header override or transport option. |
| TLS signaling | P3 | Depends on sipgo TLS support. |
| SRTP media | P3 + Custom | Not in diago. |
| ICE/STUN/TURN | P3 + Custom | Not in diago. |
| DNS SRV/NAPTR | P3 | sipgo supports RFC 3263? validate. |
| 100rel/PRACK | P3 + Custom | Not in sipgo. |
| SIP MESSAGE (IM) | P3 + Custom | Needs MESSAGE support. |
| Browser audio | Dropped (desired) | Requires WebSocket audio bridge. |

## Conclusion

The Go plan is sound for a minimal SIP phone, but several core differentiators
from `Docs/DESIGN.md` (especially browser audio and deep SIP trace tooling)
require non-trivial custom implementation beyond sipgo/diago. If those features
remain priorities, the plan should explicitly budget time for custom SIP/media
engineering rather than treating them as minor extensions.
