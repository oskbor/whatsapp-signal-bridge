package signal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"net/http"

	"github.com/gorilla/websocket"
	"github.com/oskbor/bridge/logging"
)

type Attachment struct {
	ContentType string `json:"contentType"`
	Filename    string `json:"filename"`
	Id          string `json:"id"`
	Size        int64  `json:"size"`
}

type ReceivedMessage struct {
	Envelope struct {
		Source       string
		SourceNumber string
		SourceUuid   string
		SourceName   string
		SourceDevice int
		Timestamp    int64
		DataMessage  struct {
			Timestamp        int64
			Message          string
			ExpiresInSeconds int
			ViewOnce         bool
			Mentions         []interface{}
			Attachments      []Attachment
			Contacts         []interface{}
			GroupInfo        struct {
				GroupId string
				Type    string
			}
			Destination       string
			DestinationNumber string
			DestinationUuid   string
		} `json:"dataMessage,omitempty"`
	}
}

func (message *ReceivedMessage) IsEmpty() bool {
	return message.Envelope.DataMessage.Timestamp == 0
}

type Client struct {
	channel chan ReceivedMessage
	config  *config
}

func NewClient(options ...Option) (*Client, error) {
	channel := make(chan ReceivedMessage)
	config := &config{
		Logger: logging.DefaultLogger("signal-wrapper"),
	}
	for _, option := range options {
		option(config)
	}
	config.Logger.Debug().Msgf("using host %s", config.Host)
	config.Logger.Debug().Msgf("using number %s", config.Number)

	connection := Client{
		channel: channel,
		config:  config,
	}
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://"+config.Host+"/v1/receive/"+config.Number, nil)
	if err != nil {
		return nil, err
	}
	config.Logger.Info().Msgf("Listening for signal messages on %s", conn.RemoteAddr())
	go func() {
		defer conn.Close()
		defer close(channel)
		for {
			messageType, bytes, err := conn.ReadMessage()
			if err != nil {
				panic(err)
			}
			if messageType == websocket.TextMessage {
				config.Logger.Debug().Msg("got message" + string(bytes))
				messageStruct := ReceivedMessage{}
				err := json.Unmarshal(bytes, &messageStruct)
				if err != nil {
					panic(err)
				}
				if !messageStruct.IsEmpty() {
					channel <- messageStruct
				} else {
					config.Logger.Debug().Msg("the message seems empty, ignoring")
				}

			} else {
				panic("Unexpected message type received on websocket")
			}
		}
	}()
	return &connection, nil
}

func (c *Client) OnMessage(f func(ReceivedMessage)) {
	go func() {
		for {
			message := <-c.channel
			f(message)
		}
	}()
}

type sendMessageBody struct {
	Base64Attachments []string `json:"base64_attachments"`
	Message           string   `json:"message"`
	Number            string   `json:"number"`
	Recipients        []string `json:"recipients"`
}

func (c *Client) SendMessage(message string, recipients, base64attachments []string) error {
	body, err := json.Marshal(sendMessageBody{
		Base64Attachments: base64attachments,
		Message:           message,
		Number:            c.config.Number,
		Recipients:        recipients,
	})
	if err != nil {
		return nil
	}
	res, err := http.Post("http://"+c.config.Host+"/v2/send", "application/json", bytes.NewReader(body))
	if res.StatusCode != 201 {
		respBody, _ := io.ReadAll(res.Body)
		defer res.Body.Close()
		message := res.Status
		if len(respBody) > 0 {
			message += ": " + string(respBody)
		}
		return fmt.Errorf("error sending message: %s", message)
	}
	return err
}
