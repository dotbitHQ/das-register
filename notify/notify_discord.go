package notify

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"time"
)

type Webhook struct {
	Username  string  `json:"username"`
	AvatarURL string  `json:"avatar_url"`
	Content   string  `json:"content"`
	Embeds    []Embed `json:"embeds"`
}
type Embed struct {
	Author      Author  `json:"author"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	Description string  `json:"description"`
	Color       int64   `json:"color"`
	Fields      []Field `json:"fields"`
	Thumbnail   Image   `json:"thumbnail"`
	Image       Image   `json:"image"`
	Footer      Footer  `json:"footer"`
}

type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IconURL string `json:"icon_url"`
}

type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}

type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

type Image struct {
	URL string `json:"url"`
}

func SendNotifyDiscord(webhook, content string) error {
	if webhook == "" {
		return nil
	}
	var data Webhook
	data.Content = content
	resp, _, errs := gorequest.New().Post(webhook).SendStruct(&data).Timeout(time.Second * 10).End()
	if len(errs) > 0 {
		return fmt.Errorf("errs:%v", errs)
	} else if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("http code:%d", resp.StatusCode)
	}
	return nil
}
