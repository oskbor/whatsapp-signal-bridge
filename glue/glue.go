package glue

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/oskbor/bridge/signal"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
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
		if len(text) == 0 && len(attachments) == 0 {
			fmt.Printf("Empty message, skipping %+v\n", msg)
		}

		err = g.si.SendMessage(msg.Info.PushName+": "+text, []string{groupId}, attachments)
		if err != nil {
			g.OnError(err)
			return
		}

	default:
		fmt.Printf("Received an unhandled event! \n\n %+v\n\n", msg)
	}

}
func (g *Glue) onSignalMessage(message signal.ReceivedMessage) {
	fmt.Printf("got signal message %+v\n\n", message)

	whatsappConversation, err := g.store.GetWhatsAppConversationId(message.Envelope.DataMessage.GroupInfo.GroupId)
	if err != nil {
		g.OnError(err)
		return
	}
	jid, err := types.ParseJID(whatsappConversation)
	if err != nil {
		g.OnError(err)
		return
	}
	waTextMessage := &waProto.Message{
		Conversation: &message.Envelope.DataMessage.Message,
	}
	_, err = g.wa.SendMessage(jid, "", waTextMessage)
	if err != nil {
		g.OnError(err)
		return
	}
	for _, attachment := range message.Envelope.DataMessage.Attachments {
		waMessage := &waProto.Message{}
		bytes, err := g.si.DownloadAttachment(attachment.Id)
		if err != nil {
			g.OnError(err)
			return
		}

		if strings.HasPrefix(attachment.ContentType, "image/") {
			resp, err := g.wa.Upload(context.Background(), bytes, whatsmeow.MediaImage)
			if err != nil {
				g.OnError(err)
				return
			}
			waImageMessage := &waProto.ImageMessage{
				Mimetype:      proto.String(attachment.ContentType),
				Url:           &resp.URL,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
			}
			waMessage.ImageMessage = waImageMessage

		} else if strings.HasPrefix(attachment.ContentType, "video/") {
			resp, err := g.wa.Upload(context.Background(), bytes, whatsmeow.MediaVideo)
			if err != nil {
				g.OnError(err)
				return
			}
			waVideoMessage := &waProto.VideoMessage{
				Mimetype:      proto.String(attachment.ContentType),
				Url:           &resp.URL,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
			}
			waMessage.VideoMessage = waVideoMessage

		} else if strings.HasPrefix(attachment.ContentType, "audio/") {
			resp, err := g.wa.Upload(context.Background(), bytes, whatsmeow.MediaAudio)
			if err != nil {
				g.OnError(err)
				return
			}
			waAudioMessage := &waProto.AudioMessage{
				Mimetype:      proto.String(attachment.ContentType),
				Url:           &resp.URL,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
			}
			waMessage.AudioMessage = waAudioMessage

		} else {
			resp, err := g.wa.Upload(context.Background(), bytes, whatsmeow.MediaDocument)
			if err != nil {
				g.OnError(err)
				return
			}
			waFileMessage := &waProto.DocumentMessage{
				Mimetype:      proto.String(attachment.ContentType),
				Url:           &resp.URL,
				DirectPath:    &resp.DirectPath,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
			}
			waMessage.DocumentMessage = waFileMessage
		}

		g.wa.SendMessage(jid, "", waMessage)
	}

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
			return "", fmt.Errorf("failed to create signal group: %w", err)
		}
		err = g.store.LinkGroups(conversationId.String(), signalGroupId)
		if err != nil {
			return "", fmt.Errorf("failed to link groups: %w", err)
		}
	} else if err != nil {
		return "", err
	}
	return signalGroupId, err
}
