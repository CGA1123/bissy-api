package robert_test

import (
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/robert"
)

func freshQuery(now time.Time) *robert.Query {
	oneHourAgo := now.Add(-time.Hour)
	twoHoursAgo := now.Add(-2 * time.Hour)

	return &robert.Query{
		LastRefresh: oneHourAgo,
		UpdatedAt:   twoHoursAgo,
		Lifetime:    robert.Duration(3 * time.Hour)}
}

func TestFresh(t *testing.T) {
	t.Parallel()

	now := time.Now()
	var query *robert.Query

	// fresh query
	query = freshQuery(now)
	expect.True(t, query.Fresh(now))

	// query updated after last refresh
	query = freshQuery(now)
	query.UpdatedAt = query.LastRefresh.Add(time.Second)
	expect.False(t, query.Fresh(now))

	// last refresh longer than lifetime ago
	query = freshQuery(now)
	query.LastRefresh = now.Add(-time.Duration(query.Lifetime)).Add(-time.Second)
	expect.False(t, query.Fresh(now))

	// last refresh longer than lifetime ago and updated after last refresh
	query = freshQuery(now)
	query.UpdatedAt = query.LastRefresh.Add(time.Second)
	query.LastRefresh = now.Add(-time.Duration(query.Lifetime)).Add(-time.Second)
	expect.False(t, query.Fresh(now))
}
