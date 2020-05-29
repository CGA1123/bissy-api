package robert_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/cga1123/bissy-api/expect"
	"github.com/cga1123/bissy-api/robert"
	"github.com/google/uuid"
)

func TestInMemoryCreate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestStore(now, id)
	createQuery := robert.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: 3 * robert.Duration(time.Hour),
	}

	expected := robert.Query{
		Id:          id,
		Query:       "SELECT 1;",
		Lifetime:    3 * robert.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	query, err := store.Create(&createQuery)

	expect.Ok(t, err)
	expect.Equal(t, expected, *query)
	expect.Equal(t, expected, store.Queries[id])
}

func TestInMemoryCreateSmoke(t *testing.T) {
	t.Parallel()

	store := robert.NewInMemoryStore(&robert.RealClock{}, &robert.UUIDGenerator{})
	createQuery := robert.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: 3 * robert.Duration(time.Hour),
	}

	_, err := store.Create(&createQuery)

	expect.Ok(t, err)
}

func TestInMemoryGet(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestStore(now, id)
	createQuery := robert.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: 3 * robert.Duration(time.Hour),
	}

	expected := robert.Query{
		Id:          id,
		Query:       "SELECT 1;",
		Lifetime:    3 * robert.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	_, err := store.Create(&createQuery)
	expect.Ok(t, err)

	// when id is found
	query, err := store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	// when id is not found
	_, err = store.Get("")
	expect.Error(t, err)
}

func TestInMemoryDelete(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestStore(now, id)
	createQuery := robert.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: 3 * robert.Duration(time.Hour),
	}

	expected := robert.Query{
		Id:          id,
		Query:       "SELECT 1;",
		Lifetime:    3 * robert.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	_, err := store.Create(&createQuery)
	expect.Ok(t, err)

	query, err := store.Delete(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)
	expect.Equal(t, map[string]robert.Query{}, store.Queries)

	_, err = store.Delete(id)
	expect.Error(t, err)
}

func TestInMemoryUpdate(t *testing.T) {
	t.Parallel()

	now := time.Now()
	id := uuid.New().String()
	store := newTestStore(now, id)
	createQuery := robert.CreateQuery{
		Query:    "SELECT 1;",
		Lifetime: 3 * robert.Duration(time.Hour),
	}

	newQuery := "SELECT 2;"
	newLifetime := robert.Duration(time.Hour)
	updateQuery := robert.UpdateQuery{
		Query:    &newQuery,
		Lifetime: &newLifetime,
	}

	expected := robert.Query{
		Id:          id,
		Query:       "SELECT 2;",
		Lifetime:    robert.Duration(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	_, err := store.Create(&createQuery)
	expect.Ok(t, err)

	// Test returned query
	query, err := store.Update(id, &updateQuery)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	// Test persistence
	query, err = store.Get(id)
	expect.Ok(t, err)
	expect.Equal(t, expected, *query)

	// Test partial update
	newQuery = "SELECT 15;"
	query, err = store.Update(id, &robert.UpdateQuery{Query: &newQuery})
	expect.Ok(t, err)
	expect.Equal(t, newQuery, query.Query)
	expect.Equal(t, newLifetime, query.Lifetime)

	newLifetime = 15 * robert.Duration(time.Second)
	query, err = store.Update(id, &robert.UpdateQuery{Lifetime: &newLifetime})
	expect.Ok(t, err)
	expect.Equal(t, newLifetime, query.Lifetime)
	expect.Equal(t, newQuery, query.Query)

	// Test updating lastrefresh
	query, err = store.Update(id, &robert.UpdateQuery{LastRefresh: now.Add(time.Hour)})
	expect.Ok(t, err)
	expect.Equal(t, newQuery, query.Query)
	expect.Equal(t, newLifetime, query.Lifetime)
	expect.Equal(t, now.Add(time.Hour), query.LastRefresh)

	// Updating not existing query
	_, err = store.Update("", &updateQuery)
	expect.Error(t, err)

}

func selectsFromSlice(queries []robert.Query) []string {
	selects := []string{}
	for _, query := range queries {
		selects = append(selects, query.Query)
	}

	return selects
}

func TestInMemoryList(t *testing.T) {
	t.Parallel()

	store := robert.NewInMemoryStore(&robert.RealClock{}, &robert.UUIDGenerator{})
	selects := []string{}

	for i := 0; i < 10; i++ {
		s := fmt.Sprintf("SELECT %v;", i)
		q := &robert.CreateQuery{
			Query:    s,
			Lifetime: robert.Duration(time.Duration(i) * time.Hour),
		}
		selects = append(selects, s)
		_, err := store.Create(q)

		expect.Ok(t, err)
	}

	_, err := store.List(0, 1)
	expect.Error(t, err)

	queries, err := store.List(1, 10)
	expect.Ok(t, err)
	expect.Equal(t, selects, selectsFromSlice(queries))

	queries, err = store.List(2, 3)
	expect.Ok(t, err)
	expect.Equal(t, selects[3:6], selectsFromSlice(queries))

	queries, err = store.List(4, 3)
	expect.Ok(t, err)
	expect.Equal(t, selects[9:10], selectsFromSlice(queries))

	queries, err = store.List(10, 3)
	expect.Ok(t, err)
	expect.Equal(t, []robert.Query{}, queries)

	queries, err = store.List(1, 30)
	expect.Ok(t, err)
	expect.Equal(t, selects, selectsFromSlice(queries))
}
