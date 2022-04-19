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
	return []string{base64.StdEncoding.EncodeToString(data)}
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
	case msg.Message.ExtendedTextMessage != nil && msg.Message.ExtendedTextMessage.Text != nil:
		text := *msg.Message.ExtendedTextMessage.Text
		if msg.Message.ExtendedTextMessage.Description != nil {
			text += "\n" + *msg.Message.ExtendedTextMessage.Description
		}
		return text
	case msg.Message.ImageMessage != nil && msg.Message.ImageMessage.Caption != nil:
		return *msg.Message.ImageMessage.Caption
	case msg.Message.VideoMessage != nil && msg.Message.VideoMessage.Caption != nil:
		return *msg.Message.VideoMessage.Caption
	case msg.Message.AudioMessage != nil:
		return ""
	case msg.Message.DocumentMessage != nil && msg.Message.DocumentMessage.Title != nil:
		return *msg.Message.DocumentMessage.Title
	case msg.Message.StickerMessage != nil:
		return ""
	default:
		return ""
	}
}
