package querycache_test

import (
	"testing"
	"time"

	"github.com/cga1123/bissy-api/querycache"
	"github.com/cga1123/bissy-api/utils/expect"
)

func testSharelinkCreate(t *testing.T, query querycache.Query, store querycache.SharelinkStore, id string, now time.Time) {
	datasource, err := datasourceStore.Create(userID, &querycache.CreateDatasource{})
	expect.Ok(t, err)

	createQuery := querycache.CreateQuery{
		Query:        "SELECT 1;",
		Lifetime:     3 * querycache.Duration(time.Hour),
		DatasourceID: datasource.ID,
	}

	expected := &querycache.Query{
		ID:           id,
		UserID:       userID,
		Query:        "SELECT 1;",
		DatasourceID: datasource.ID,
		Lifetime:     3 * querycache.Duration(time.Hour),
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRefresh:  now,
	}

	query, err := store.Create(userID, &createQuery)

	expect.Ok(t, err)
	expect.Equal(t, expected, query)

	query, err = store.Get(userID, id)
	expect.Ok(t, err)
	expect.Equal(t, expected, query)
}
