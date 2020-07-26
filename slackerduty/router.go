package slackerduty

import (
	"fmt"
	"io"
	"net/http"

	"github.com/PagerDuty/go-pagerduty"
	"github.com/cga1123/bissy-api/utils/handlerutils"
	"github.com/gorilla/mux"
)

const baseURL = "https://app.pagerduty.com"

// Config contains the configuration for slackerduty
type Config struct {
	PagerdutyWebhookToken string
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

	fmt.Println(messages)

	return nil
}

func parseWebhook(body io.ReadCloser) (*pagerduty.WebhookPayloadMessages, error) {
	if body == nil {
		return nil, &handlerutils.HandlerError{
			Err: fmt.Errorf("empty body"), Status: http.StatusBadRequest}
	}

	return pagerduty.DecodeWebhook(body)
}
