package tui

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rivo/tview"
	"github.com/siptty/siptty/internal/engine"
)

const maxTraceMessages = 1000

// TracePanel displays a scrolling SIP message trace log.
// Trace events are buffered and flushed to the tview.TextView periodically
// to avoid blocking the eventLoop goroutine with tview's synchronous draw calls.
type TracePanel struct {
	view     *tview.TextView
	msgCount int

	mu      sync.Mutex
	pending strings.Builder
}

// NewTracePanel creates a scrolling TextView for SIP trace messages.
func NewTracePanel() *TracePanel {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tv.SetTitle("SIP Trace").SetBorder(true)
	return &TracePanel{view: tv}
}

// Buffer formats a SIP trace event and appends it to the pending buffer.
// Goroutine-safe. Does not touch tview widgets directly.
func (p *TracePanel) Buffer(ev engine.SipTraceEvent) {
	var color, arrow string
	switch ev.Direction {
	case "send":
		color = "green"
		arrow = "→"
	case "recv":
		color = "yellow"
		arrow = "←"
	default:
		color = "white"
		arrow = "?"
	}

	ts := ev.Timestamp.Format("15:04:05.000")
	firstLine := firstSIPLine(ev.Message)

	entry := fmt.Sprintf("[%s]%s %s %s %s %s %s[-]\n[%s]%s[-]\n",
		color, ts, arrow, ev.Transport, ev.LocalAddr, arrow, ev.RemoteAddr,
		color, firstLine,
	)

	p.mu.Lock()
	p.pending.WriteString(entry)
	p.mu.Unlock()
}

// Flush writes all pending trace text to the tview.TextView and scrolls to end.
// Must be called on the tview main goroutine (inside QueueUpdateDraw).
func (p *TracePanel) Flush() {
	p.mu.Lock()
	text := p.pending.String()
	p.pending.Reset()
	p.mu.Unlock()

	if text == "" {
		return
	}

	fmt.Fprint(p.view, text)
	p.msgCount += strings.Count(text, "\n") / 2

	if p.msgCount > maxTraceMessages {
		p.trim()
	}

	p.view.ScrollToEnd()
}

// trim removes the oldest messages to keep the buffer at maxTraceMessages.
func (p *TracePanel) trim() {
	text := p.view.GetText(false)
	lines := strings.Split(text, "\n")
	excess := p.msgCount - maxTraceMessages
	linesToRemove := excess * 2
	if linesToRemove >= len(lines) {
		p.view.Clear()
		p.msgCount = 0
		return
	}
	p.view.Clear()
	fmt.Fprint(p.view, strings.Join(lines[linesToRemove:], "\n"))
	p.msgCount = maxTraceMessages
}

// firstSIPLine extracts the first non-empty line from a SIP message.
func firstSIPLine(msg string) string {
	for _, line := range strings.SplitN(msg, "\n", 2) {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return "(empty)"
}
