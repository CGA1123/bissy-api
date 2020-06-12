package querycache

import (
	"fmt"

	"github.com/cga1123/bissy-api/utils"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

type SQLAdapterStore struct {
	db          *hnysqlx.DB
	clock       utils.Clock
	idGenerator utils.IdGenerator
}

func NewSQLAdapterStore(db *hnysqlx.DB, clock utils.Clock, generator utils.IdGenerator) *SQLAdapterStore {
	return &SQLAdapterStore{db: db, clock: clock, idGenerator: generator}
}

func (s *SQLAdapterStore) Create(ca *CreateAdapter) (*Adapter, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	query := `
		INSERT INTO querycache_adapters (id, name, type, options, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`

	var adapter Adapter
	if err := s.db.Get(&adapter, query, id, ca.Name, ca.Type, ca.Options, now, now); err != nil {
		return nil, err
	}

	return &adapter, nil
}

func (s *SQLAdapterStore) Get(id string) (*Adapter, error) {
	var adapter Adapter

	query := "SELECT * FROM querycache_adapters WHERE id = $1"

	if err := s.db.Get(&adapter, query, id); err != nil {
		return nil, err
	}

	return &adapter, nil
}

func (s *SQLAdapterStore) Delete(id string) (*Adapter, error) {
	var adapter Adapter

	query := "DELETE FROM querycache_adapters WHERE id = $1 RETURNING *"

	if err := s.db.Get(&adapter, query, id); err != nil {
		return nil, err
	}

	return &adapter, nil
}

func (s *SQLAdapterStore) List(page, per int) ([]*Adapter, error) {
	if page < 1 || per < 1 {
		return nil,
			fmt.Errorf("page and per must be greater than 0 (page %v) (per %v)",
				page, per)
	}

	adapters := []*Adapter{}

	query := `SELECT * FROM querycache_adapters ORDER BY created_at OFFSET $1 LIMIT $2`
	if err := s.db.Select(&adapters, query, (page-1)*per, per); err != nil {
		return nil, err
	}

	return adapters, nil
}

func (s *SQLAdapterStore) Update(id string, ua *UpdateAdapter) (*Adapter, error) {
	var adapter Adapter

	query := `
		UPDATE querycache_adapters
		SET name = COALESCE($2, name),
				type = COALESCE($3, type),
				options = COALESCE($4, options)
		WHERE id = $1
		RETURNING *`

	if err := s.db.Get(&adapter, query, id, ua.Name, ua.Type, ua.Options); err != nil {
		return nil, err
	}

	return &adapter, nil
}
