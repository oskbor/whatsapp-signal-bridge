package glue

import (
	"database/sql"
	"fmt"

	"github.com/oskbor/bridge/signal"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

type Glue struct {
	wa    *whatsmeow.Client
	si    *signal.Client
	store *Store
	cfg   *config
}

func (g *Glue) onWhatsAppEvent(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		// find which conversation this message belongs to
		// check if corresponding signal conversation exists
		// if not create, then send the message
		fmt.Printf("Received a message! \n\n %+v\n\n", v)
	default:
		fmt.Printf("Received an unhandled event! \n\n %+v\n\n", v)
	}

}
func (g *Glue) onSignalMessage(message signal.ReceivedMessage) {
	fmt.Println("got message", message)

}

func New(whatsmeow *whatsmeow.Client, si *signal.Client, options ...Option) *Glue {
	cfg := &config{}
	for _, option := range options {
		option(cfg)
	}
	if cfg.SignalRecipient == "" {
		panic("SignalRecipient is required")
	}
	db, err := sql.Open("sqlite3", "file:glue.db?_foreign_keys=on")
	if err != nil {
		panic(fmt.Errorf("failed to open database: %w", err))
	}
	store, err := NewStore(db)
	if err != nil {
		panic(fmt.Errorf("failed to create store: %w", err))
	}
	g := &Glue{
		wa:    whatsmeow,
		si:    si,
		store: store,
		cfg:   cfg,
	}

	g.wa.AddEventHandler(g.onWhatsAppEvent)
	g.si.OnMessage(g.onSignalMessage)
	return g

}
