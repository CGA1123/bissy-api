package ping

import (
	"encoding/json"
	"net/http"

	"github.com/cga1123/bissy-api/handlerutils"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	handlerutils.ContentType(w, handlerutils.ContentTypeJson)

	json.NewEncoder(w).Encode(map[string]string{"message": "pong"})
}
