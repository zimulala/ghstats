package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// TitleColor defines feishu message title color.
type TitleColor string

const (
	// TitleColorBlue Blue.
	TitleColorBlue TitleColor = "blue"
	// TitleColorWathet Wathet
	TitleColorWathet TitleColor = "wathet"
	// TitleColorTurquoise Turquoise
	TitleColorTurquoise TitleColor = "turquoise"
	// TitleColorGreen Green
	TitleColorGreen TitleColor = "green"
	// TitleColorYellow Yellow
	TitleColorYellow TitleColor = "yellow"
	// TitleColorOrange Orange
	TitleColorOrange TitleColor = "orange"
	// TitleColorRed Red
	TitleColorRed TitleColor = "red"
	// TitleColorCarmine Carmine
	TitleColorCarmine TitleColor = "carmine"
	// TitleColorViolet Violet
	TitleColorViolet TitleColor = "violet"
	// TitleColorPurple Purple
	TitleColorPurple TitleColor = "purple"
	// TitleColorIndigo Indigo
	TitleColorIndigo TitleColor = "indigo"
	// TitleColorGrey Grey
	TitleColorGrey TitleColor = "grey"
)

// WebhookBot is a feishu webhook bot.
type WebhookBot struct {
	Token  string
	IsTest bool // If it's true, we only print the message to local.
}

// SendMarkdownMessage sends markdown message via feishu bot,
// msg must be markdown escaped.
//
//	{
//	  "msg_type": "interactive",
//	  "card": {
//	    "config": {
//	      "wide_screen_mode": true,
//	      "enable_forward": true
//	    },
//	    "elements": [
//	      {
//	        "tag": "div",
//	        "text": {
//	          "content": "",
//	          "tag": "lark_md"
//	        }
//	      }
//	    ],
//	    "header": {
//	      "title": {
//	        "content": "PTAL ❤️",
//	        "tag": "plain_text"
//	      }
//	    }
//	  }
//	}
//
// Source: https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN#4996824a
func (bot WebhookBot) SendMarkdownMessage(ctx context.Context, title, msg string, titleColor TitleColor) error {
	msg1, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	payload := fmt.Sprintf(`
	{
		"msg_type": "interactive",
		"card": {
			"config": {
				"wide_screen_mode": true,
				"enable_forward": true
			},
			"header": {
				"title": {
					"tag": "plain_text",
					"content": %s
				},
				"template":"%s"
			},
			"elements": [
				{
					"tag": "div",
					"text": {
						"tag": "lark_md",
						"content": %s
					}
				}
			]
		}
	}`, title, titleColor, string(msg1))
	if bot.IsTest {
		fmt.Printf("Print messages locally only: %s\n", payload)
		return nil
	}

	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", bot.Token)
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, url, strings.NewReader(payload))
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("feishu send markdown error [%d] %s", resp.StatusCode, string(body))
	}
	return err
}
