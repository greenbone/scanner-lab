package sink

import (
	"fmt"

	"github.com/greenbone/scanner-lab/feature-tests/featuretest"
	"github.com/mattermost/mattermost-server/v5/model"
)

type Mattermost struct {
	client    *model.Client4
	user      *model.User
	channelID string
}

func NewMattermost(address, channelID, token string) (*Mattermost, error) {
	client := model.NewAPIv4Client(address)
	client.SetToken(token)
	return &Mattermost{
		client:    client,
		channelID: channelID,
	}, nil

}

func (m *Mattermost) Send(results []featuretest.Result) error {
	post := model.Post{
		ChannelId: m.channelID,
		Message:   ":information_desk_person: **scanner-lab** results:",
	}
	p, resp := m.client.CreatePost(&post)
	if resp.Error != nil {
		return resp.Error
	}
	post.RootId = p.Id
	for _, r := range results {
		if r.FailureDescription != "" {
			post.Message = fmt.Sprintf(":warning: **%s**: failed; %s; `%s`", r.Name, r.Duration, r.FailureDescription)
		} else {
			post.Message = fmt.Sprintf(":white_check_mark: **%s**: succeded; %s", r.Name, r.Duration)
		}
		_, resp = m.client.CreatePost(&post)
		if resp.Error != nil {
			return resp.Error
		}
	}

	return nil
}

func (m *Mattermost) Error(err error) error {

	post := model.Post{
		ChannelId: m.channelID,
		Message:   ":interrobang: **scanner-lab** failed with error:",
	}
	p, resp := m.client.CreatePost(&post)
	if resp.Error != nil {
		return resp.Error
	}
	post.RootId = p.Id
	post.Message = fmt.Sprintf("```\n%s```", err.Error())
	_, resp = m.client.CreatePost(&post)
	return resp.Error
}
