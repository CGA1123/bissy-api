package robert

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type IdGenerator interface {
	Generate() string
}

type UUIDGenerator struct{}

func (generator *UUIDGenerator) Generate() string {
	return uuid.New().String()
}

type RealClock struct{}

func (clock *RealClock) Now() time.Time {
	return time.Now()
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
	query, ok := store.Queries[id]
	if !ok {
		return nil, fmt.Errorf("Query (id: %v) not found", id)
	}

	return &query, nil
}

func (store *InMemoryStore) Delete(id string) (*Query, error) {
	query, ok := store.Queries[id]
	if !ok {
		return nil, fmt.Errorf("Query (id: %v) not found", id)
	}

	delete(store.Queries, id)

	return &query, nil
}

func (store *InMemoryStore) Update(id string, update *UpdateQuery) (*Query, error) {
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
