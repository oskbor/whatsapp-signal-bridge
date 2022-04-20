package glue

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/oskbor/bridge/logging"
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

		groupId, err := g.GetOrCreateSignalGroup(msg)
		if err != nil {
			g.OnError(err)
			return
		}
		isFromMe := msg.Info.MessageSource.IsFromMe
		if isFromMe {
			return
		}
		text := g.ExtractTextContent(msg)
		attachments := g.ExtractAttachments(msg)
		if len(text) == 0 && len(attachments) == 0 {
			g.cfg.Logger.Warn().Msgf("Empty message, skipping %+v\n", msg)
			return
		}
		err = g.si.SendMessage(msg.Info.PushName+": "+text, []string{groupId}, attachments)
		if err != nil {
			g.OnError(err)
			return
		}
	}

}
func (g *Glue) onSignalMessage(message signal.ReceivedMessage) {
	g.cfg.Logger.Debug().Msgf("got signal message %+v\n", message)

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
	cfg := &config{
		Logger: logging.DefaultLogger("glue"),
	}
	for _, option := range options {
		option(cfg)
	}
	cfg.Logger.Debug().Msgf("using recipient %s", cfg.SignalRecipient)
	if cfg.SignalRecipient == "" {
		panic("SignalRecipient is required")
	}
	db, err := sql.Open("sqlite3", "file:./bridge/glue.db?_foreign_keys=on")
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
	g.cfg.Logger.Error().Msg(e.Error())
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
		waChatName = waChatName + " on WhatsApp"
		g.cfg.Logger.Debug().Msg("Creating signal group " + waChatName)
		signalGroupId, err = g.si.CreateGroup(waChatName, SIGNAL_GROUP_DESCRIPTION, signal.Disabled, []string{g.cfg.SignalRecipient}, signal.OnlyAdmins, signal.EveryMember)
		if err != nil {
			return "", fmt.Errorf("failed to create signal group: %w", err)
		}
		g.cfg.Logger.Debug().Msg("successfully created group with ID " + signalGroupId)
		info, err := g.si.GetGroupInfo(signalGroupId)
		if err != nil {
			return "", fmt.Errorf("failed to get signal group info: %w", err)
		}
		g.cfg.Logger.Debug().Msg("successfully fetched internal id for group " + info.InternalId)
		err = g.store.LinkGroups(conversationId.String(), signalGroupId, info.InternalId)
		if err != nil {
			return "", fmt.Errorf("failed to link groups: %w", err)
		}
		g.cfg.Logger.Debug().Msg("link successful")
	} else if err != nil {
		return "", err
	}
	return signalGroupId, err
}
