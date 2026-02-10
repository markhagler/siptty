package tui

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"
	"github.com/siptty/siptty/internal/engine"
)

const maxTraceMessages = 1000

// TracePanel displays a scrolling SIP message trace log.
type TracePanel struct {
	view     *tview.TextView
	msgCount int
}

// NewTracePanel creates a scrolling TextView for SIP trace messages.
func NewTracePanel() *TracePanel {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true)
	tv.SetTitle("SIP Trace").SetBorder(true)
	return &TracePanel{view: tv}
}

// Append formats and appends a SIP trace event to the view.
// Send messages are green (→), received messages are yellow (←).
// Keeps at most maxTraceMessages entries by trimming from the front.
func (p *TracePanel) Append(ev engine.SipTraceEvent) {
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

	fmt.Fprint(p.view, entry)
	p.msgCount++

	// Trim old messages if we exceed the limit.
	if p.msgCount > maxTraceMessages {
		p.trim()
	}

	p.view.ScrollToEnd()
}

// trim removes the oldest messages to keep the buffer at maxTraceMessages.
func (p *TracePanel) trim() {
	text := p.view.GetText(false)
	lines := strings.Split(text, "\n")
	// Each message is 2 lines (header + first SIP line), trim oldest entries.
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
