package tui

import (
	"fmt"

	"github.com/rivo/tview"
	"github.com/siptty/siptty/internal/engine"
)

// AccountPanel displays SIP account registration state.
type AccountPanel struct {
	list     *tview.List
	accounts map[string]int // accountID -> list index
}

// NewAccountPanel creates a tview.List with title "ACCOUNTS" and border.
func NewAccountPanel() *AccountPanel {
	list := tview.NewList().
		ShowSecondaryText(true)
	list.SetTitle("ACCOUNTS").SetBorder(true)
	return &AccountPanel{
		list:     list,
		accounts: make(map[string]int),
	}
}

// Update processes a RegStateEvent and updates the account display.
// Colored bullet: green "●" registered, red "○" unregistered, yellow "◉" failed.
func (p *AccountPanel) Update(ev engine.RegStateEvent) {
	var bullet string
	switch ev.State {
	case "registered":
		bullet = "[green]●[-]"
	case "unregistered":
		bullet = "[red]○[-]"
	case "failed":
		bullet = "[yellow]◉[-]"
	default:
		bullet = "[grey]?[-]"
	}

	primary := fmt.Sprintf("%s %s", bullet, ev.AccountID)
	secondary := fmt.Sprintf("  %s", ev.State)
	if ev.Reason != "" {
		secondary = fmt.Sprintf("  %s (%s)", ev.State, ev.Reason)
	}

	if idx, ok := p.accounts[ev.AccountID]; ok {
		p.list.SetItemText(idx, primary, secondary)
	} else {
		idx := p.list.GetItemCount()
		p.list.AddItem(primary, secondary, 0, nil)
		p.accounts[ev.AccountID] = idx
	}
}
