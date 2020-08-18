package slackerduty

import (
	"fmt"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/slack-go/slack"
)

func IncidentBlock(msg *pagerduty.WebhookPayload) slack.Block {
	incident := msg.Incident
	title := fmt.Sprintf("*<%v|%v>*", incident.HTMLURL, IncidentTitle(msg))

	return slack.NewSectionBlock(
		slack.NewTextBlockObject(slack.MarkdownType, title, false, false),
		[]*slack.TextBlockObject{},
		nil,
	)
}

func IncidentTitle(msg *pagerduty.WebhookPayload) string {
	return fmt.Sprintf("*[%v] %v*", msg.Incident.ID, msg.Incident.Title)
}

func AddBlock(blocks []slack.Block, message *pagerduty.WebhookPayload, blockFn func(*pagerduty.WebhookPayload) slack.Block) []slack.Block {
	if block := blockFn(message); block != nil {
		return append(blocks, block)
	}

	return blocks
}

func BuildBlocks(message *pagerduty.WebhookPayload) []slack.Block {
	blocks := []slack.Block{}
	blocks = AddBlock(blocks, message, IncidentBlock)

	return blocks
}
