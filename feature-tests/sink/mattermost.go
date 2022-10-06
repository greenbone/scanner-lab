package sink

import (
	"encoding/xml"
	"fmt"

	"github.com/google/uuid"
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

func (m *Mattermost) SendSingleResult(r featuretest.Result) error {

	post := model.Post{
		ChannelId: m.channelID,
	}
	var fids []string
	if r.Resp != nil {
		fmt.Printf("sending resp\n")
		xr, err := xml.MarshalIndent(r.Resp, "", "")
		if err == nil {
			fup, up := m.client.UploadFile(xr, m.channelID, fmt.Sprintf("%s-%s.xml", r.Name, uuid.NewString()))
			if up.Error == nil {
				fids = make([]string, len(fup.FileInfos))
				for i, fi := range fup.FileInfos {
					fids[i] = fi.Id
				}
			}
		}
	}
	post.FileIds = fids
	if r.FailureDescription != "" {
		post.Message = fmt.Sprintf(":warning: **%s**: failed; %s; `%s`", r.Name, r.Duration, r.FailureDescription)
	} else {
		post.Message = fmt.Sprintf(":white_check_mark: **%s**: succeded; %s", r.Name, r.Duration)
	}
	_, resp := m.client.CreatePost(&post)
	return resp.Error
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
		if err := m.SendSingleResult(r); err != nil {
			return err
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
