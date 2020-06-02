package querycache

import (
	"github.com/jmoiron/sqlx"
)

type SQLAdapterStore struct {
	db          *sqlx.DB
	clock       Clock
	idGenerator IdGenerator
}

func NewSQLAdapterStore(db *sqlx.DB, clock Clock, generator IdGenerator) *SQLAdapterStore {
	return &SQLAdapterStore{db: db, clock: clock, idGenerator: generator}
}

func (s *SQLAdapterStore) Create(ca *CreateAdapter) (*Adapter, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	var adapter Adapter
	err := s.db.QueryRowx(`
		INSERT INTO querycache_adapters (id, name, type, options, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`, id, ca.Name, ca.Type, ca.Options, now, now).StructScan(&adapter)
	if err != nil {
		return nil, err
	}

	return &adapter, nil
}

func (s *SQLAdapterStore) Get(id string) (*Adapter, error) {
	var adapter Adapter
	err := s.db.QueryRowx("SELECT * FROM querycache_adapters WHERE id = $1", id).StructScan(&adapter)
	if err != nil {
		return nil, err
	}

	return &adapter, nil
}

func (s *SQLAdapterStore) Delete(id string) (*Adapter, error) {
	var adapter Adapter
	err := s.db.QueryRowx("DELETE FROM querycache_adapters WHERE id = $1 RETURNING *", id).StructScan(&adapter)
	if err != nil {
		return nil, err
	}

	return &adapter, nil
}
