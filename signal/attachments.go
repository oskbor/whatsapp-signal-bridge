package signal

import (
	"fmt"
	"io"
	"net/http"
)

func (c *Client) DownloadAttachment(id string) ([]byte, error) {
	res, err := http.Get("http://" + c.config.Host + "/v1/attachments/" + id)
	if err != nil {
		return nil, err
	}
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error downloading attachment: %s", res.Status)
	}
	return respBody, err
}
