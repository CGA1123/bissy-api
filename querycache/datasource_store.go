package querycache

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/cga1123/bissy-api/utils"
)

// Datasource describes a database that Queries may be related to and executed
// against
type Datasource struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Options   string    `json:"options"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// UpdateDatasource describes the paramater which may be updated on a Datasource
type UpdateDatasource struct {
	Name    *string `json:"name"`
	Type    *string `json:"type"`
	Options *string `json:"options"`
}

// CreateDatasource describes the required paramater to create a new Datasource
type CreateDatasource struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Options string `json:"options"`
}

// DatasourceStore describes a generic Store for Datasources
type DatasourceStore interface {
	Get(string) (*Datasource, error)
	Create(*CreateDatasource) (*Datasource, error)
	List(page int, per int) ([]*Datasource, error)
	Delete(string) (*Datasource, error)
	Update(string, *UpdateDatasource) (*Datasource, error)
}

// InMemoryDatasourceStore implements an in-memory DatasourceStore
type InMemoryDatasourceStore struct {
	Store       map[string]*Datasource
	clock       utils.Clock
	idGenerator utils.IDGenerator
	lock        sync.RWMutex
}

// NewExecutor returns a new SQLExecutor configured against this Datasource
// Will return a TestExecutor if the datasources "Type" is "test"
func (a *Datasource) NewExecutor() (Executor, error) {
	switch a.Type {
	case "test":
		return &TestExecutor{}, nil
	default:
		// TODO: Cache this per datasource-id? keep DB objects available and not need to recreate connections?
		return NewSQLExecutor(a.Type, a.Options)
	}
}

// NewInMemoryDatasourceStore returns a new InMemoryDatasourceStore
func NewInMemoryDatasourceStore(clock utils.Clock, idGenerator utils.IDGenerator) *InMemoryDatasourceStore {
	return &InMemoryDatasourceStore{
		clock:       clock,
		idGenerator: idGenerator,
		Store:       map[string]*Datasource{},
	}
}

// Create creates and persists a new Datasource to the Store
func (s *InMemoryDatasourceStore) Create(a *CreateDatasource) (*Datasource, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	id := s.idGenerator.Generate()
	datasource := Datasource{
		ID:        id,
		Name:      a.Name,
		Type:      a.Type,
		Options:   a.Options,
		CreatedAt: s.clock.Now(),
		UpdatedAt: s.clock.Now(),
	}

	s.Store[id] = &datasource

	return &datasource, nil
}

// Delete removes the Datasource with given id from the Store
func (s *InMemoryDatasourceStore) Delete(id string) (*Datasource, error) {
	datasource, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.Store, id)

	return datasource, nil
}

// Get returns the Datasource with associated id from the store
func (s *InMemoryDatasourceStore) Get(id string) (*Datasource, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	datasource, ok := s.Store[id]
	if !ok {
		return nil, fmt.Errorf("Datasource (id: %v) not found", id)
	}

	return datasource, nil
}

// Update updates the Datasource with associated id from the store
func (s *InMemoryDatasourceStore) Update(id string, u *UpdateDatasource) (*Datasource, error) {
	datasource, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if u.Name != nil {
		datasource.Name = *u.Name
	}

	if u.Type != nil {
		datasource.Type = *u.Type
	}

	if u.Options != nil {
		datasource.Options = *u.Options
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	return datasource, nil
}

// List returns the requests Datasources from the Store, ordered by createdAt
func (s *InMemoryDatasourceStore) List(page, per int) ([]*Datasource, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if page < 1 || per < 1 {
		return nil,
			fmt.Errorf("page and per must be greater than 0 (page %v) (per %v)",
				page, per)
	}

	datasources := []*Datasource{}
	for _, datasource := range s.Store {
		datasources = append(datasources, datasource)
	}

	sort.Slice(
		datasources,
		func(i, j int) bool {
			timeI := datasources[i].CreatedAt
			timeJ := datasources[j].CreatedAt

			return timeI.Before(timeJ)
		},
	)

	datasourceCount := len(s.Store)
	startIndex := (page - 1) * per
	endIndex := startIndex + per

	if startIndex > datasourceCount-1 {
		return []*Datasource{}, nil
	}

	if endIndex > datasourceCount {
		endIndex = datasourceCount
	}

	return datasources[startIndex:endIndex], nil
}
