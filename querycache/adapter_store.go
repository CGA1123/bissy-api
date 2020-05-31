package querycache

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type Adapter struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Options   string    `json:"options"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateAdapter struct {
	Name    *string `json:"name"`
	Type    *string `json:"type"`
	Options *string `json:"options"`
}

type CreateAdapter struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Options string `json:"options"`
}

type AdapterStore interface {
	Get(string) (*Adapter, error)
	Create(*CreateAdapter) (*Adapter, error)
	List(page int, per int) ([]Adapter, error)
	Delete(string) (*Adapter, error)
	Update(string, *UpdateAdapter) (*Adapter, error)
}

type InMemoryAdapterStore struct {
	Store       map[string]Adapter
	clock       Clock
	idGenerator IdGenerator
	lock        sync.RWMutex
}

func (a *Adapter) NewExecutor() (Executor, error) {
	switch a.Type {
	case "test":
		return &TestExecutor{}, nil
	default:
		// TODO: Cache this per adapter-id? keep DB objects available and not need to recreate connections?
		return NewSQLExecutor(a.Type, a.Options)
	}
}

func NewInMemoryAdapterStore(clock Clock, idGenerator IdGenerator) *InMemoryAdapterStore {
	return &InMemoryAdapterStore{
		clock:       clock,
		idGenerator: idGenerator,
		Store:       map[string]Adapter{},
	}
}

func (s *InMemoryAdapterStore) Create(a *CreateAdapter) (*Adapter, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	id := s.idGenerator.Generate()
	adapter := Adapter{
		Id:        id,
		Name:      a.Name,
		Type:      a.Type,
		Options:   a.Options,
		CreatedAt: s.clock.Now(),
		UpdatedAt: s.clock.Now(),
	}

	s.Store[id] = adapter

	return &adapter, nil
}

func (s *InMemoryAdapterStore) Delete(id string) (*Adapter, error) {
	adapter, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.Store, id)

	return adapter, nil
}

func (s *InMemoryAdapterStore) Get(id string) (*Adapter, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	adapter, ok := s.Store[id]
	if !ok {
		return nil, fmt.Errorf("Adapter (id: %v) not found", id)
	}

	return &adapter, nil
}

func (s *InMemoryAdapterStore) Update(id string, u *UpdateAdapter) (*Adapter, error) {
	adapter, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	if u.Name != nil {
		adapter.Name = *u.Name
	}

	if u.Type != nil {
		adapter.Type = *u.Type
	}

	if u.Options != nil {
		adapter.Options = *u.Options
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.Store[id] = *adapter

	return adapter, nil
}

func (s *InMemoryAdapterStore) List(page, per int) ([]Adapter, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0 (is %v)", page)
	}

	adapters := []Adapter{}
	for _, adapter := range s.Store {
		adapters = append(adapters, adapter)
	}

	sort.Slice(
		adapters,
		func(i, j int) bool {
			timeI := adapters[i].CreatedAt
			timeJ := adapters[j].CreatedAt

			return timeI.Before(timeJ)
		},
	)

	adapterCount := len(s.Store)
	startIndex := (page - 1) * per
	endIndex := startIndex + per

	if startIndex > adapterCount-1 {
		return []Adapter{}, nil
	}

	if endIndex > adapterCount {
		endIndex = adapterCount
	}

	return adapters[startIndex:endIndex], nil
}
