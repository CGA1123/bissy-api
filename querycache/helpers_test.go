package querycache_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/cga1123/bissy-api/auth"
	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

func newTestQueryStore(db *hnysqlx.DB, now time.Time, id string) *querycache.SQLQueryStore {
	return querycache.NewSQLQueryStore(
		db,
		&utils.TestClock{Time: now},
		&utils.TestIDGenerator{ID: id})
}

func newTestDatasourceStore(db *hnysqlx.DB, now time.Time, id string) *querycache.SQLDatasourceStore {
	return querycache.NewSQLDatasourceStore(
		db,
		&utils.TestClock{Time: now},
		&utils.TestIDGenerator{ID: id})
}

func testClaims() *auth.Claims {
	return &auth.Claims{UserID: uuid.New().String()}
}

func testConfig(db *hnysqlx.DB) (time.Time, string, *querycache.Config) {
	now := time.Now().Truncate(time.Millisecond)
	id := uuid.New().String()

	return now, id, &querycache.Config{
		QueryStore:      newTestQueryStore(db, now, id),
		DatasourceStore: newTestDatasourceStore(db, now, id),
		Executor:        &querycache.TestExecutor{}}
}

func testHandler(claims *auth.Claims, c *querycache.Config, r *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()

	router := mux.NewRouter()
	router.Use(auth.TestMiddleware(claims))
	c.SetupHandlers(router)

	router.ServeHTTP(recorder, r)

	return recorder
}

func jsonBody(v interface{}) (*bytes.Reader, error) {
	body, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(body), nil
}
