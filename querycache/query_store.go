package querycache

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"
)

type Query struct {
	Id          string    `json:"id"`
	Query       string    `json:"query"`
	Lifetime    Duration  `json:"lifetime"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	LastRefresh time.Time `json:"lastRefresh"`
}

func (query *Query) Fresh(now time.Time) bool {
	timeSinceLastRefresh := now.Sub(query.LastRefresh)
	refreshedRecently := Duration(timeSinceLastRefresh) < query.Lifetime
	updatedRecently := query.UpdatedAt.After(query.LastRefresh)

	return refreshedRecently && !updatedRecently
}

type CreateQuery struct {
	Query    string   `json:"query"`
	Lifetime Duration `json:"lifetime"`
}

type UpdateQuery struct {
	Query       *string   `json:"query"`
	Lifetime    *Duration `json:"lifetime"`
	LastRefresh time.Time `json:"lastRefresh"`
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

type QueryStore interface {
	Get(string) (*Query, error)
	Create(*CreateQuery) (*Query, error)
	List(page int, per int) ([]Query, error)
	Delete(string) (*Query, error)
	Update(string, *UpdateQuery) (*Query, error)
}

type InMemoryStore struct {
	Queries     map[string]Query
	clock       Clock
	idGenerator IdGenerator
	lock        sync.RWMutex
}

func NewInMemoryStore(clock Clock, idGenerator IdGenerator) *InMemoryStore {
	return &InMemoryStore{
		clock:       clock,
		idGenerator: idGenerator,
		Queries:     map[string]Query{},
	}
}

// Create and persist to memory a Query from a CreateQuery struct
func (store *InMemoryStore) Create(createQuery *CreateQuery) (*Query, error) {
	id := store.idGenerator.Generate()
	now := store.clock.Now()

	query := Query{
		Id:          id,
		Query:       createQuery.Query,
		Lifetime:    createQuery.Lifetime,
		CreatedAt:   now,
		UpdatedAt:   now,
		LastRefresh: now,
	}

	store.lock.Lock()
	defer store.lock.Unlock()

	store.Queries[id] = query

	return &query, nil
}

func (store *InMemoryStore) Get(id string) (*Query, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	query, ok := store.Queries[id]
	if !ok {
		return nil, fmt.Errorf("Query (id: %v) not found", id)
	}

	return &query, nil
}

func (store *InMemoryStore) Delete(id string) (*Query, error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	query, ok := store.Queries[id]
	if !ok {
		return nil, fmt.Errorf("Query (id: %v) not found", id)
	}

	delete(store.Queries, id)

	return &query, nil
}

func (store *InMemoryStore) Update(id string, update *UpdateQuery) (*Query, error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	query, ok := store.Queries[id]
	if !ok {
		return nil, fmt.Errorf("Query (id: %v) not found", id)
	}

	if update.Query != nil {
		query.Query = *update.Query
	}

	if update.Lifetime != nil {
		query.Lifetime = *update.Lifetime
	}

	if !update.LastRefresh.IsZero() {
		query.LastRefresh = update.LastRefresh
	}

	store.Queries[id] = query

	return &query, nil
}

func (store *InMemoryStore) List(page, per int) ([]Query, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0 (is %v)", page)
	}

	queries := []Query{}
	for _, query := range store.Queries {
		queries = append(queries, query)
	}

	sort.Slice(
		queries,
		func(i, j int) bool {
			timeI := queries[i].CreatedAt
			timeJ := queries[j].CreatedAt

			return timeI.Before(timeJ)
		},
	)

	queryCount := len(store.Queries)
	startIndex := (page - 1) * per
	endIndex := startIndex + per

	if startIndex > queryCount-1 {
		return []Query{}, nil
	}

	if endIndex > queryCount {
		endIndex = queryCount
	}

	return queries[startIndex:endIndex], nil
}
