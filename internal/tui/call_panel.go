package tui

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/siptty/siptty/internal/engine"
)

// callRow tracks a call's position in the table.
type callRow struct {
	row   int
	state string
}

// CallPanel displays active calls and a dial input.
type CallPanel struct {
	table     *tview.Table
	dialInput *tview.InputField
	flex      *tview.Flex
	calls     map[string]*callRow
	nextRow   int
}

// NewCallPanel creates a call table and dial input wrapped in a vertical Flex.
func NewCallPanel() *CallPanel {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(true, false)
	table.SetTitle("CALLS").SetBorder(true)

	// Header row.
	headers := []string{"ID", "Remote", "State", "Duration"}
	for col, h := range headers {
		table.SetCell(0, col, tview.NewTableCell("[bold]"+h+"[-]").
			SetSelectable(false).
			SetExpansion(1))
	}

	dialInput := tview.NewInputField().
		SetLabel("Dial: ").
		SetFieldWidth(40)

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(table, 0, 1, true).
		AddItem(dialInput, 1, 0, false)

	return &CallPanel{
		table:     table,
		dialInput: dialInput,
		flex:      flex,
		calls:     make(map[string]*callRow),
		nextRow:   1, // row 0 is header
	}
}

// Update processes a CallStateEvent and adds or updates a call row.
func (p *CallPanel) Update(ev engine.CallStateEvent) {
	cr, ok := p.calls[ev.CallID]
	if !ok {
		cr = &callRow{row: p.nextRow}
		p.calls[ev.CallID] = cr
		p.nextRow++
	}
	cr.state = ev.State

	color := stateColor(ev.State)
	row := cr.row

	p.table.SetCell(row, 0, tview.NewTableCell(ev.CallID).SetTextColor(color))
	p.table.SetCell(row, 1, tview.NewTableCell(ev.RemoteURI).SetTextColor(color))
	p.table.SetCell(row, 2, tview.NewTableCell(ev.State).SetTextColor(color))
	p.table.SetCell(row, 3, tview.NewTableCell(formatDuration(ev.Duration)).SetTextColor(color))
}

// ShowDTMF briefly highlights the DTMF digit on the call row.
func (p *CallPanel) ShowDTMF(ev engine.DTMFEvent) {
	cr, ok := p.calls[ev.CallID]
	if !ok {
		return
	}
	// Append DTMF indicator to the state cell.
	cell := p.table.GetCell(cr.row, 2)
	cell.SetText(fmt.Sprintf("%s [%c]", cr.state, ev.Digit))
}

// SelectedCallID returns the call ID of the currently selected table row.
func (p *CallPanel) SelectedCallID() string {
	row, _ := p.table.GetSelection()
	if row < 1 || row >= p.nextRow {
		return ""
	}
	cell := p.table.GetCell(row, 0)
	if cell == nil {
		return ""
	}
	return cell.Text
}

func stateColor(state string) tcell.Color {
	switch state {
	case "confirmed":
		return tcell.ColorGreen
	case "ringing", "early", "incoming":
		return tcell.ColorYellow
	case "disconnected":
		return tcell.ColorRed
	case "calling":
		return tcell.ColorDarkCyan
	default:
		return tcell.ColorWhite
	}
}

func formatDuration(d time.Duration) string {
	s := int(d.Seconds())
	if s < 0 {
		return "0:00"
	}
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}
