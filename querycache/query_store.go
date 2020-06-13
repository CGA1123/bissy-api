package querycache

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/cga1123/bissy-api/utils"
)

// Query describes an SQL query on a given datasource that should be cached for
// a given Lifetime value
type Query struct {
	ID           string    `json:"id"`
	Query        string    `json:"query"`
	DatasourceID string    `json:"datasourceId" db:"datasource_id"`
	Lifetime     Duration  `json:"lifetime"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	LastRefresh  time.Time `json:"lastRefresh" db:"last_refresh"`
}

// Fresh determines whether a query was last refreshed within Lifetime of the
// given time parameter
func (query *Query) Fresh(now time.Time) bool {
	timeSinceLastRefresh := now.Sub(query.LastRefresh)
	refreshedRecently := Duration(timeSinceLastRefresh) < query.Lifetime
	updatedRecently := query.UpdatedAt.After(query.LastRefresh)

	return refreshedRecently && !updatedRecently
}

// CreateQuery describes the required parameter to create a new Query
type CreateQuery struct {
	Query        string   `json:"query"`
	Lifetime     Duration `json:"lifetime"`
	DatasourceID string   `json:"datasourceId"`
}

// UpdateQuery describes the paramater which may be updated on a Query
type UpdateQuery struct {
	Query        *string   `json:"query"`
	DatasourceID *string   `json:"datasourceId"`
	Lifetime     *Duration `json:"lifetime"`
	LastRefresh  time.Time `json:"lastRefresh"`
}

// Duration is an alias to time.Duration to allow for defining JSON marshalling
// and unmarshalling
type Duration time.Duration

// MarshalJSON marshals a Duration into JSON
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON unmarshals JSON into a Duration
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

// QueryStore describes a generic Store for Queries
type QueryStore interface {
	Get(string) (*Query, error)
	Create(*CreateQuery) (*Query, error)
	List(page int, per int) ([]*Query, error)
	Delete(string) (*Query, error)
	Update(string, *UpdateQuery) (*Query, error)
}

// InMemoryQueryStore defines an in-memory implementation of a QueryStore
type InMemoryQueryStore struct {
	Queries     map[string]*Query
	clock       utils.Clock
	idGenerator utils.IDGenerator
	lock        sync.RWMutex
}

// NewInMemoryQueryStore builds a new InMemoryQueryStore
func NewInMemoryQueryStore(clock utils.Clock, idGenerator utils.IDGenerator) *InMemoryQueryStore {
	return &InMemoryQueryStore{
		clock:       clock,
		idGenerator: idGenerator,
		Queries:     map[string]*Query{},
	}
}

// Create creates and persist to memory a Query from a CreateQuery struct
func (store *InMemoryQueryStore) Create(createQuery *CreateQuery) (*Query, error) {
	id := store.idGenerator.Generate()
	now := store.clock.Now()

	query := Query{
		ID:           id,
		Query:        createQuery.Query,
		Lifetime:     createQuery.Lifetime,
		DatasourceID: createQuery.DatasourceID,
		CreatedAt:    now,
		UpdatedAt:    now,
		LastRefresh:  now,
	}

	store.lock.Lock()
	defer store.lock.Unlock()

	store.Queries[id] = &query

	return &query, nil
}

// Get returns the Query with associated id from the store
func (store *InMemoryQueryStore) Get(id string) (*Query, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	query, ok := store.Queries[id]
	if !ok {
		return nil, fmt.Errorf("Query (id: %v) not found", id)
	}

	return query, nil
}

// Delete removes the Query with associated id from the store
func (store *InMemoryQueryStore) Delete(id string) (*Query, error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	query, ok := store.Queries[id]
	if !ok {
		return nil, fmt.Errorf("Query (id: %v) not found", id)
	}

	delete(store.Queries, id)

	return query, nil
}

// Update updates the Query with associated id from the store
func (store *InMemoryQueryStore) Update(id string, update *UpdateQuery) (*Query, error) {
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

	if update.DatasourceID != nil {
		query.DatasourceID = *update.DatasourceID
	}

	if !update.LastRefresh.IsZero() {
		query.LastRefresh = update.LastRefresh
	}

	return query, nil
}

// List returns the requests Queries from the Store, ordered by createdAt
func (store *InMemoryQueryStore) List(page, per int) ([]*Query, error) {
	store.lock.RLock()
	defer store.lock.RUnlock()

	if page < 1 || per < 1 {
		return nil,
			fmt.Errorf("page and per must be greater than 0 (page %v) (per %v)",
				page, per)
	}

	queries := []*Query{}
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
		return []*Query{}, nil
	}

	if endIndex > queryCount {
		endIndex = queryCount
	}

	return queries[startIndex:endIndex], nil
}
