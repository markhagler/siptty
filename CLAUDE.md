# AGENTS.md

This guide helps future coding assistants produce Go code that is readable, idiomatic, and aligned with the style used in `sipgo` and `diago` (Emiago).

# Documentation
-When creating or modifying documentation our standard documentation folder name is Docs/ with a capital D. Always use this directory when looking for documentation, writing documentation, or updating documentation.


## Scope
- Applies to this repository (`/Users/hagler/code/go_test`) and subdirectories.
- Prioritize clarity, correctness, and maintainability over clever abstractions.
- For VoIP/SIP code, lifecycle correctness (context, cleanup, dialog/transaction state) is more important than micro-optimizations.

## Go Project Best Practices
- Keep code simple and explicit.
- Use `gofmt`/`goimports` formatting conventions consistently.
- Prefer short functions with one clear responsibility.
- Use early returns for errors; avoid deep nesting.
- Wrap errors with context using `%w` (`fmt.Errorf("...: %w", err)`).
- Use `errors.Join` when returning multiple cleanup/close errors.
- Keep interfaces small and close to consumers.
- Avoid global mutable state.

## Recommended Repository Organization
- `cmd/<app>/main.go`: executable entry points.
- `internal/<domain>/...`: app-specific business logic.
- `pkg/...`: only for truly reusable public libraries.
- `*_test.go` colocated with implementation files for unit/integration tests.
- Optional: `configs/`, `scripts/`, `docs/` as needed.

## Style Alignment With Emiago (`sipgo`/`diago`)

### Naming and API Shape
- Prefer constructor + options pattern:
  - `NewX(...)`
  - `WithX(...)` option functions
- Use direct, call-flow oriented method names for SIP/dialog actions:
  - `Trying`, `Ringing`, `Answer`, `Invite`, `WaitAnswer`, `Hangup`, `Close`.
- Keep receiver names short and conventional:
  - `c`, `d`, `srv`, `ua`.
- Prefer explicit protocol naming (`udp`, `tcp`, `tls`, `ws`, `wss`) and normalize transport strings.

### Error Handling
- Propagate errors immediately with context.
- Keep error text specific to operation and object (`request`, `dialog`, `transport`, etc.).
- For cleanup methods, make `Close()` idempotent and safe to call more than once.

### Context and Lifecycle
- Pass `context.Context` as first argument for blocking/network operations.
- Tie goroutines and network listeners to context cancellation.
- Ensure transaction/dialog/media resources are terminated/closed on all paths (`defer` where appropriate).
- In SIP flows, be explicit about ACK, CANCEL, and transaction termination behavior.

### Logging
- Use structured logging (`log/slog`) with stable keys.
- Preferred keys include: `error`, `req.method`, `addr`, `protocol`, `id`.
- Log useful state transitions (start/stop/listen/handle/fail), not noisy internals.

### Comments and Documentation
- Keep comments short and practical.
- Add comments for protocol caveats and behavioral contracts (thread safety, ACK requirements, experimental APIs).
- Avoid redundant comments that restate obvious code.

## VoIP/SIP-Specific Guardrails
- Treat SIP signaling and RTP/media as explicit state machines.
- Be cautious with thread safety when mutating invite/dialog/session state.
- Keep transport/address/contact handling explicit for NAT and multi-interface cases.
- Prefer deterministic cleanup and timeout behavior over implicit background handling.

## Testing Guidance
- Add tests for:
  - Dialog state transitions
  - Transaction success/failure paths
  - Timeouts/cancellation
  - Media/session setup and teardown
- Favor readable integration-style tests for call flows and protocol behavior.

## tview TUI Pitfalls (Lessons Learned)
- **`tview.Table` infinite loop**: `SetSelectable(true, ...)` with zero selectable cells causes `Table.InputHandler.forward()` to spin forever. Start tables with `SetSelectable(false, false)` and enable only when data rows exist.
- **All tview draw calls block**: `QueueUpdateDraw`, `QueueUpdate`, and `Draw` all block the calling goroutine until the main loop processes them. Never call these from a high-frequency event path.
- **Buffer + debounce for fast events**: For high-volume updates (SIP trace, logs), buffer into a `strings.Builder` with a mutex and flush on a 50ms timer via a single `QueueUpdateDraw`. This keeps the event-reading goroutine free-running.
- **Profile before guessing**: Use `-debug` flag with `net/http/pprof` on `:6060` to get CPU profiles and goroutine dumps. The real bottleneck is often not where you think — our 100% CPU was a tview bug, not SIP event throughput.
- **Read library source on blocking semantics**: tview docs say `Draw()` is "thread-safe" but don't mention it blocks. Always check the actual implementation of third-party APIs for hidden synchronization.

## Non-Goals
- Do not introduce heavy architectural patterns unless justified by complexity.
- Do not over-generalize early; keep APIs practical and incremental.

## Quick Do/Don't Examples
- Do: use constructor+options APIs (`NewClient(..., WithClientHostname(...))`); Don't: add many positional args or config booleans to one function.
- Do: pass `context.Context` first for network/dialog/media operations; Don't: start long-running goroutines without cancellation wiring.
- Do: return wrapped errors (`fmt.Errorf("invite failed: %w", err)`); Don't: return raw errors with no operation context.
- Do: make `Close()` idempotent and aggregate cleanup errors (`errors.Join`); Don't: assume cleanup is called once or ignore close failures.
- Do: use structured `slog` fields (`"error", err, "req.method", ...`); Don't: log unstructured strings that are hard to filter.
- Do: keep SIP call-flow methods explicit (`Trying`, `Ringing`, `Answer`, `Hangup`); Don't: hide protocol steps behind vague method names.
