package tui

import (
	"fmt"
	"log/slog"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/siptty/siptty/internal/engine"
)

const (
	focusBorderColor  = tcell.ColorAqua
	normalBorderColor = tcell.ColorWhite
)

// EngineInterface abstracts the engine so the TUI depends only on event types.
type EngineInterface interface {
	Events() <-chan engine.Event
	Accounts() []string
	Dial(accountID, uri string) error
	Answer(callID string) error
	Hangup(callID string) error
	SendDTMF(callID string, digit rune) error
	Transfer(callID, target string) error
	PlayAudio(callID, path string) error
}

// App is the top-level TUI application.
type App struct {
	app      *tview.Application
	engine   EngineInterface
	accounts *AccountPanel
	calls    *CallPanel
	trace    *TracePanel
	blf      *tview.Table
	dialogs  *tview.TextView
	pages    *tview.Pages
	grid     *tview.Grid

	// focusable panels for Tab cycling
	panels []tview.Primitive
	focus  int

	// overlay prevents global keys from interfering with modals/prompts
	overlay bool
}

// NewApp builds the full tview layout and returns an App.
func NewApp(eng EngineInterface) *App {
	a := &App{
		app:    tview.NewApplication(),
		engine: eng,
	}

	// Build panels.
	a.accounts = NewAccountPanel()
	a.calls = NewCallPanel()
	a.trace = NewTracePanel()
	a.blf = newBLFPlaceholder()
	a.dialogs = newDialogsPlaceholder()

	// Bottom tabbed section — trace and dialogs only (calls live in the top row).
	a.pages = tview.NewPages().
		AddPage("trace", a.trace.view, true, true).
		AddPage("dialogs", a.dialogs, true, false)

	tabBar := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]1[white]:SIP Trace  [yellow]2[white]:SIP Dialogs")

	// Header.
	title := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[bold]siptty")
	headerRight := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignRight).
		SetText("[grey]?:Help  F10:Quit")
	header := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(title, 0, 1, false).
		AddItem(headerRight, 0, 1, false)

	// Footer.
	footer := tview.NewTextView().
		SetDynamicColors(true).
		SetText("[yellow]d[white]:Dial [yellow]a[white]:Ans [yellow]h[white]:Hang [yellow]x[white]:Xfer [yellow]p[white]:DTMF [yellow]Tab[white]:Focus [yellow]1/2[white]:Tabs [yellow]?[white]:Help")

	// Top three-column row.
	topRow := tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.accounts.list, 0, 1, false).
		AddItem(a.calls.flex, 0, 2, true).
		AddItem(a.blf, 0, 1, false)

	// Bottom section with tab bar and pages.
	bottomSection := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(tabBar, 1, 0, false).
		AddItem(a.pages, 0, 1, false)

	// Main grid.
	a.grid = tview.NewGrid().
		SetRows(1, 0, 1, -1, 1).
		SetColumns(0).
		AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(topRow, 1, 0, 1, 1, 0, 0, true).
		AddItem(tview.NewBox().SetBorder(false), 2, 0, 1, 1, 0, 0, false).
		AddItem(bottomSection, 3, 0, 1, 1, 0, 0, false).
		AddItem(footer, 4, 0, 1, 1, 0, 0, false)

	// Wire up dial input: Enter dials, Escape returns focus.
	a.calls.dialInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			uri := a.calls.dialInput.GetText()
			if uri != "" {
				accounts := a.engine.Accounts()
				if len(accounts) == 0 {
					a.setStatus("No accounts configured")
				} else if err := a.engine.Dial(accounts[0], uri); err != nil {
					a.setStatus(fmt.Sprintf("Dial error: %v", err))
				}
				a.calls.dialInput.SetText("")
			}
			a.app.SetFocus(a.calls.table)
			a.focus = 1
			a.highlightFocus()
		case tcell.KeyEscape:
			a.calls.dialInput.SetText("")
			a.app.SetFocus(a.calls.table)
			a.focus = 1
			a.highlightFocus()
		}
	})

	// Focus cycle: top-row panels only.
	a.panels = []tview.Primitive{
		a.accounts.list,
		a.calls.table,
		a.blf,
	}
	a.focus = 1 // start on calls table
	a.highlightFocus()

	a.setupKeyBindings()

	return a
}

// Run starts the event loop goroutine and runs the tview application.
func (a *App) Run() error {
	a.app.SetRoot(a.grid, true)
	go a.eventLoop()
	return a.app.Run()
}

// Stop gracefully stops the tview application.
func (a *App) Stop() {
	a.app.Stop()
}

// eventLoop reads engine events and dispatches to the appropriate panels.
func (a *App) eventLoop() {
	ch := a.engine.Events()
	if ch == nil {
		return
	}
	for ev := range ch {
		switch e := ev.(type) {
		case engine.RegStateEvent:
			a.app.QueueUpdateDraw(func() {
				a.accounts.Update(e)
			})
		case engine.CallStateEvent:
			a.app.QueueUpdateDraw(func() {
				a.calls.Update(e)
			})
		case engine.SipTraceEvent:
			// Write directly to tview.TextView (goroutine-safe via io.Writer)
			// and trigger a non-blocking draw. This avoids QueueUpdateDraw which
			// blocks the caller and can backpressure the SIP stack.
			a.trace.Append(e)
			a.app.Draw()
		case engine.DTMFEvent:
			a.app.QueueUpdateDraw(func() {
				a.calls.ShowDTMF(e)
			})
		}
	}
}

func (a *App) setupKeyBindings() {
	a.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// Quit always works, even over overlays.
		switch event.Key() {
		case tcell.KeyCtrlC, tcell.KeyF10:
			a.app.Stop()
			return nil
		}

		// While an overlay (help, transfer prompt, DTMF prompt) is active,
		// let the overlay handle all other input.
		if a.overlay {
			return event
		}

		// If dial input is focused, let it handle its own keys (except Escape).
		if a.app.GetFocus() == a.calls.dialInput {
			if event.Key() == tcell.KeyEscape {
				a.app.SetFocus(a.calls.table)
				a.focus = 1
				a.highlightFocus()
				return nil
			}
			return event
		}

		switch event.Key() {
		case tcell.KeyF1:
			a.showHelp()
			return nil
		case tcell.KeyTab:
			a.cycleFocus()
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case '?':
				a.showHelp()
				return nil
			case 'd':
				a.app.SetFocus(a.calls.dialInput)
				return nil
			case 'a':
				a.answerSelected()
				return nil
			case 'h':
				a.hangupSelected()
				return nil
			case 'x':
				a.promptTransfer()
				return nil
			case 'p':
				a.promptDTMF()
				return nil
			case '1':
				a.pages.SwitchToPage("trace")
				return nil
			case '2':
				a.pages.SwitchToPage("dialogs")
				return nil
			}
		}
		return event
	})
}

func (a *App) cycleFocus() {
	a.focus = (a.focus + 1) % len(a.panels)
	a.app.SetFocus(a.panels[a.focus])
	a.highlightFocus()
}

// highlightFocus colors the focused panel's border and resets the others.
func (a *App) highlightFocus() {
	for i, p := range a.panels {
		color := normalBorderColor
		if i == a.focus {
			color = focusBorderColor
		}
		setBorderColor(p, color)
	}
}

func setBorderColor(p tview.Primitive, color tcell.Color) {
	switch v := p.(type) {
	case *tview.List:
		v.SetBorderColor(color)
	case *tview.Table:
		v.SetBorderColor(color)
	case *tview.TextView:
		v.SetBorderColor(color)
	}
}

// restoreGrid resets the root to the main grid and restores focus after an overlay.
func (a *App) restoreGrid() {
	a.overlay = false
	a.app.SetRoot(a.grid, true)
	a.app.SetFocus(a.panels[a.focus])
	a.highlightFocus()
}

func (a *App) answerSelected() {
	callID := a.calls.SelectedCallID()
	if callID == "" {
		return
	}
	if err := a.engine.Answer(callID); err != nil {
		a.setStatus(fmt.Sprintf("Answer error: %v", err))
	}
}

func (a *App) hangupSelected() {
	callID := a.calls.SelectedCallID()
	if callID == "" {
		return
	}
	if err := a.engine.Hangup(callID); err != nil {
		a.setStatus(fmt.Sprintf("Hangup error: %v", err))
	}
}

func (a *App) promptTransfer() {
	callID := a.calls.SelectedCallID()
	if callID == "" {
		return
	}
	a.overlay = true
	input := tview.NewInputField().
		SetLabel("Transfer to: ")
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			target := input.GetText()
			if target != "" {
				_ = a.engine.Transfer(callID, target)
			}
		}
		a.restoreGrid()
	})
	a.app.SetRoot(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.grid, 0, 1, false).
		AddItem(input, 1, 0, true),
		true,
	)
	a.app.SetFocus(input)
}

func (a *App) promptDTMF() {
	callID := a.calls.SelectedCallID()
	if callID == "" {
		return
	}
	a.overlay = true
	input := tview.NewInputField().
		SetLabel("DTMF digits: ").
		SetAcceptanceFunc(func(text string, lastChar rune) bool {
			return (lastChar >= '0' && lastChar <= '9') || lastChar == '*' || lastChar == '#'
		})
	input.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			for _, digit := range input.GetText() {
				_ = a.engine.SendDTMF(callID, digit)
			}
		}
		a.restoreGrid()
	})
	a.app.SetRoot(tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(a.grid, 0, 1, false).
		AddItem(input, 1, 0, true),
		true,
	)
	a.app.SetFocus(input)
}

func (a *App) showHelp() {
	a.overlay = true
	modal := tview.NewModal().
		SetText("siptty — SIP Terminal Client\n\n" +
			"NAVIGATION\n" +
			"  Tab ............ Cycle panel focus\n" +
			"  1 / 2 .......... Switch bottom tabs\n" +
			"  Escape ......... Cancel input\n\n" +
			"CALL CONTROL\n" +
			"  d .............. Dial a SIP URI\n" +
			"  a .............. Answer incoming call\n" +
			"  h .............. Hangup selected call\n" +
			"  x .............. Transfer call\n" +
			"  p .............. Send DTMF digits\n\n" +
			"GENERAL\n" +
			"  ? / F1 ......... This help\n" +
			"  F10 / Ctrl-C ... Quit").
		AddButtons([]string{"Close"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			a.restoreGrid()
		})
	a.app.SetRoot(modal, true)
}

func (a *App) setStatus(msg string) {
	slog.Info("tui status", "msg", msg)
	a.trace.view.SetTitle(fmt.Sprintf("SIP Trace — %s", msg))
}

func newBLFPlaceholder() *tview.Table {
	table := tview.NewTable().
		SetBorders(false).
		SetSelectable(false, false)
	table.SetTitle("BLF / PRESENCE").SetBorder(true)

	// Header row.
	table.SetCell(0, 0, tview.NewTableCell("[bold]Ext").SetSelectable(false))
	table.SetCell(0, 1, tview.NewTableCell("[bold]Name").SetSelectable(false))
	table.SetCell(0, 2, tview.NewTableCell("[bold]State").SetSelectable(false))

	// Placeholder rows — wired to SUBSCRIBE/NOTIFY in Phase 2.
	table.SetCell(1, 0, tview.NewTableCell("201"))
	table.SetCell(1, 1, tview.NewTableCell("Bob"))
	table.SetCell(1, 2, tview.NewTableCell("[green]●Idle[-]"))

	table.SetCell(2, 0, tview.NewTableCell("202"))
	table.SetCell(2, 1, tview.NewTableCell("Carol"))
	table.SetCell(2, 2, tview.NewTableCell("[red]◉Busy[-]"))

	return table
}

func newDialogsPlaceholder() *tview.TextView {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetText("SIP dialog viewer — available in a future release")
	tv.SetTitle("SIP Dialogs").SetBorder(true)
	return tv
}
