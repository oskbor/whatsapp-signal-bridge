package signal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GroupPermissions struct {
	AddMembers string `json:"add_members"`
	EditGroup  string `json:"edit_group"`
}
type createGroupBody struct {
	Description string           `json:"description"`
	GroupLink   string           `json:"group_link"`
	Members     []string         `json:"members"`
	Name        string           `json:"name"`
	Permissions GroupPermissions `json:"permissions"`
}

type GroupLink int

const (
	Disabled GroupLink = iota
	Enabled
	EnabledWithApproval
)

func (s GroupLink) String() string {
	stringMap := [...]string{"disabled", "enabled", "enabled-with-approval"}
	return stringMap[s]
}

type Permission int

const (
	OnlyAdmins = iota
	EveryMember
)

func (s Permission) String() string {
	stringMap := [...]string{"only-admins", "every-member"}
	return stringMap[s]
}

type createGroupResponse struct {
	Id    string `json:"id"`
	Error string `json:"error"`
}

func (c *Client) CreateGroup(name, description string, groupLink GroupLink, members []string, addMembers, EditGroup Permission) (string, error) {
	body, err := json.Marshal(createGroupBody{
		Description: description,
		GroupLink:   groupLink.String(),
		Members:     members,
		Name:        name,
		Permissions: GroupPermissions{
			AddMembers: addMembers.String(),
			EditGroup:  EditGroup.String(),
		},
	})
	if err != nil {
		return "", err
	}
	res, err := http.Post("http://"+c.config.Host+"/v1/groups/"+c.config.Number, "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 201 {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("error creating group: %s %s", res.Status, string(respBody))
	}
	payload := createGroupResponse{}
	err = json.Unmarshal(respBody, &payload)
	return payload.Id, err
}
func (c *Client) DeleteGroup(groupId string) error {

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodDelete, "http://"+c.config.Host+"/v1/groups/"+c.config.Number+"/"+groupId, nil)
	if err != nil {
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error deleting group: %s %s", resp.Status, body)
	}
	return nil
}
