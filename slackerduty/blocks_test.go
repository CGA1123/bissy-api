package slackerduty_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/cga1123/bissy-api/utils/expect"
)

func parsePayload(t *testing.T, path string) *pagerduty.WebhookPayload {
	file, err := os.Open("./payloads/" + path)
	expect.Ok(t, err)

	var payload pagerduty.WebhookPayloadMessages

	expect.Ok(t, json.NewDecoder(file).Decode(&payload))

	return &payload.Messages[0]
}

func TestIncidentBlock(t *testing.T) {
	t.Parallel()

	payload := parsePayload(t, "plain_incident_trigger.json")
}
