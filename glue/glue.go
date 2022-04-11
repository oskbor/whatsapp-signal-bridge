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

const SIGNAL_GROUP_DESCRIPTION = "Whatsapp <-> Signal bridge group"

func (g *Glue) onWhatsAppEvent(evt interface{}) {
	switch msg := evt.(type) {
	case *events.Message:
		isFromMe := msg.Info.MessageSource.IsFromMe
		if isFromMe {
			return
		}
		groupId, err := g.GetOrCreateSignalGroup(msg)
		if err != nil {
			g.OnError(err)
			return
		}
		text := g.ExtractTextContent(msg)
		attachments := g.ExtractAttachments(msg)

		err = g.si.SendMessage(text, []string{groupId}, attachments)
		if err != nil {
			g.OnError(err)
			return
		}

	default:
		fmt.Printf("Received an unhandled event! \n\n %+v\n\n", msg)
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
func (g *Glue) OnError(e error) {
	fmt.Println("[glue error]", error.Error(e))
}

func (g *Glue) GetOrCreateSignalGroup(waMessage *events.Message) (string, error) {
	conversationId := waMessage.Info.MessageSource.Chat
	signalGroupId, err := g.store.GetSignalGroupId(conversationId.String())

	if err == ErrNotFound {
		var waChatName string
		if waMessage.Info.IsGroup {
			info, err := g.wa.GetGroupInfo(conversationId)
			if err != nil {
				return "", err
			}
			waChatName = info.Name
		} else {
			info, err := g.wa.Store.Contacts.GetContact(conversationId)
			if err == nil {
				waChatName = info.FullName
			} else {
				waChatName = waMessage.Info.PushName
			}
		}
		signalGroupId, err = g.si.CreateGroup(waChatName+" on WhatsApp", SIGNAL_GROUP_DESCRIPTION, signal.Disabled, []string{g.cfg.SignalRecipient}, signal.OnlyAdmins, signal.EveryMember)
		if err != nil {
			return "", fmt.Errorf("failed to create signal group: %v\n", err)
		}
		err = g.store.LinkGroups(conversationId.String(), signalGroupId)
		if err != nil {
			return "", fmt.Errorf("failed to link groups: %v\n", err)
		}
	} else if err != nil {
		return "", err
	}
	return signalGroupId, err
}
