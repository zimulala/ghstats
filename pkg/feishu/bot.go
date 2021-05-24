package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// WebhookBot is a feishu webhook bot.
type WebhookBot string

// SendMarkdownMessage sends markdown message via feishu bot,
// msg must be markdown escaped.
//
// {
//   "msg_type": "interactive",
//   "card": {
//     "config": {
//       "wide_screen_mode": true,
//       "enable_forward": true
//     },
//     "elements": [
//       {
//         "tag": "div",
//         "text": {
//           "content": "",
//           "tag": "lark_md"
//         }
//       }
//     ],
//     "header": {
//       "title": {
//         "content": "PTAL ❤️",
//         "tag": "plain_text"
//       }
//     }
//   }
// }
//
// Source: https://open.feishu.cn/document/ukTMukTMukTM/ucTM5YjL3ETO24yNxkjN#4996824a
func (bot WebhookBot) SendMarkdownMessage(ctx context.Context, title, msg string) error {
	title1, err := json.Marshal(title)
	if err != nil {
		return err
	}
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
				}
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
	}`, string(title1), string(msg1))
	url := fmt.Sprintf("https://open.feishu.cn/open-apis/bot/v2/hook/%s", bot)
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
