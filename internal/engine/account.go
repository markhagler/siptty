package engine

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/emiago/diago"
	"github.com/emiago/sipgo/sip"
	"github.com/siptty/siptty/internal/config"
)

// Account holds registration state for a SIP account.
type Account struct {
	ID     string
	Config config.AccountConfig
	State  string // "registered", "unregistered", "failed"

	regTx  *diago.RegisterTransaction
	cancel context.CancelFunc
}

// register performs SIP registration for this account using the provided diago instance.
// It pushes RegStateEvents onto the events channel.
func (a *Account) register(ctx context.Context, dg *diago.Diago, events chan<- Event) {
	regCtx, cancel := context.WithCancel(ctx)
	a.cancel = cancel

	registrarStr := a.Config.Registrar
	var registrar sip.Uri
	if err := sip.ParseUri(registrarStr, &registrar); err != nil {
		slog.Error("invalid registrar URI", "account", a.ID, "error", err)
		a.State = "failed"
		events <- RegStateEvent{
			AccountID: a.ID,
			State:     "failed",
			Reason:    fmt.Sprintf("invalid registrar URI: %v", err),
		}
		return
	}

	expiry := time.Duration(a.Config.RegExpiry) * time.Second

	regTx, err := dg.RegisterTransaction(regCtx, registrar, diago.RegisterOptions{
		Username: a.Config.AuthUser,
		Password: a.Config.AuthPassword,
		Expiry:   expiry,
	})
	if err != nil {
		slog.Error("register transaction failed", "account", a.ID, "error", err)
		a.State = "failed"
		events <- RegStateEvent{
			AccountID: a.ID,
			State:     "failed",
			Reason:    fmt.Sprintf("register transaction: %v", err),
		}
		return
	}
	a.regTx = regTx

	if err := regTx.Register(regCtx); err != nil {
		slog.Error("registration failed", "account", a.ID, "error", err)
		a.State = "failed"
		events <- RegStateEvent{
			AccountID: a.ID,
			State:     "failed",
			Reason:    fmt.Sprintf("register: %v", err),
		}
		return
	}

	a.State = "registered"
	slog.Info("registered", "account", a.ID)
	events <- RegStateEvent{
		AccountID: a.ID,
		State:     "registered",
	}
}

// unregister sends a SIP unregistration.
func (a *Account) unregister() {
	if a.regTx != nil {
		unregCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.regTx.Unregister(unregCtx); err != nil {
			slog.Warn("unregister failed", "account", a.ID, "error", err)
		}
	}
	if a.cancel != nil {
		a.cancel()
	}
	a.State = "unregistered"
}
