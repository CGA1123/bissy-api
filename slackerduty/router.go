package slackerduty

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/cga1123/bissy-api/utils/handlerutils"
	"github.com/gorilla/mux"
	"github.com/slack-go/slack"
)

// Config contains the configuration for slackerduty
type Config struct {
	PagerdutyWebhookToken string
	SlackClient           *slack.Client
	SlackChannel          string
}

// SetupHandlers adds slackerduty HTTP handlers to the given router
func (c *Config) SetupHandlers(router *mux.Router) {
	router.
		Handle("/pagerduty/event/{token}", &handlerutils.Handler{H: c.pagerdutyEvent}).
		Methods("POST")
}

func (c *Config) pagerdutyEvent(w http.ResponseWriter, r *http.Request) error {
	token, _ := handlerutils.Params(r).Get("token")
	if token != c.PagerdutyWebhookToken {
		return &handlerutils.HandlerError{
			Status: http.StatusUnauthorized, Err: fmt.Errorf("bad token")}
	}

	messages, err := parseWebhook(r.Body)
	if err != nil {
		return err
	}

	msgsJson, _ := json.Marshal(messages)
	fmt.Println(msgsJson)

	for _, message := range messages.Messages {
		incident := message.Incident
		message := fmt.Sprintf("[%v] %v (%v)", incident.ID, incident.Title, message.Event)

		_, _, err := c.SlackClient.PostMessage(
			c.SlackChannel,
			slack.MsgOptionText(message, false))
		if err != nil {
			fmt.Printf("error: %v\n", err)
		}
	}

	return nil
}

func parseWebhook(body io.ReadCloser) (*pagerduty.WebhookPayloadMessages, error) {
	if body == nil {
		return nil, &handlerutils.HandlerError{
			Err: fmt.Errorf("empty body"), Status: http.StatusBadRequest}
	}

	return pagerduty.DecodeWebhook(body)
}
