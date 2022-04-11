package glue

import (
	"encoding/base64"

	"go.mau.fi/whatsmeow/types/events"
)

func (g *Glue) ExtractAttachments(message *events.Message) []string {

	data, err := g.wa.DownloadAny(message.RawMessage)
	if err != nil {
		return nil
	}
	return []string{base64.RawStdEncoding.EncodeToString(data)}
}

func (g *Glue) ExtractTextContent(msg *events.Message) string {
	if msg == nil {
		return ""
	}
	if msg.Message == nil {
		return ""
	}
	if msg.Message.GetConversation() != "" {
		return msg.Message.GetConversation()
	}
	switch {
	case msg.Message.ImageMessage != nil:
		return *msg.Message.ImageMessage.Caption
	case msg.Message.VideoMessage != nil:
		return *msg.Message.VideoMessage.Caption
	case msg.Message.AudioMessage != nil:
		return ""
	case msg.Message.DocumentMessage != nil:
		return *msg.Message.DocumentMessage.Title
	case msg.Message.StickerMessage != nil:
		return ""
	default:
		return ""
	}
}