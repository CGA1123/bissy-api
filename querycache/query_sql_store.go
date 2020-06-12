package querycache

import (
	"database/sql"
	"fmt"

	"github.com/cga1123/bissy-api/utils"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

type SQLQueryStore struct {
	db          *hnysqlx.DB
	clock       utils.Clock
	idGenerator utils.IdGenerator
}

func NewSQLQueryStore(db *hnysqlx.DB, clock utils.Clock, generator utils.IdGenerator) *SQLQueryStore {
	return &SQLQueryStore{db: db, clock: clock, idGenerator: generator}
}

func (s *SQLQueryStore) Get(id string) (*Query, error) {
	var query Query

	queryStr := "SELECT * FROM querycache_queries WHERE id = $1"

	if err := s.db.Get(&query, queryStr, id); err != nil {
		return nil, err
	}

	return &query, nil
}

func (s *SQLQueryStore) Create(ca *CreateQuery) (*Query, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	queryStr := `
		INSERT INTO querycache_queries (id, query, lifetime, datasource_id, created_at, updated_at, last_refresh)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *`

	var query Query
	if err := s.db.Get(&query, queryStr, id, ca.Query, ca.Lifetime, ca.DatasourceId, now, now, now); err != nil {
		return nil, err
	}

	return &query, nil
}

func (s *SQLQueryStore) Delete(id string) (*Query, error) {
	var query Query

	queryStr := "DELETE FROM querycache_queries WHERE id = $1 RETURNING *"

	if err := s.db.Get(&query, queryStr, id); err != nil {
		return nil, err
	}

	return &query, nil
}

func (s *SQLQueryStore) Update(id string, uq *UpdateQuery) (*Query, error) {
	var query Query

	queryStr := `
		UPDATE querycache_queries
		SET query = COALESCE($2, query),
				datasource_id = COALESCE($3, datasource_id),
				lifetime = COALESCE($4, lifetime),
				last_refresh = COALESCE($5, last_refresh),
				updated_at = $6
		WHERE id = $1
		RETURNING *`

	var lastRefresh sql.NullTime
	if uq.LastRefresh.IsZero() {
		lastRefresh = sql.NullTime{Valid: false}
	} else {
		lastRefresh = sql.NullTime{Time: uq.LastRefresh, Valid: true}
	}

	if err := s.db.Get(&query, queryStr, id, uq.Query, uq.DatasourceId, uq.Lifetime, lastRefresh, s.clock.Now()); err != nil {
		return nil, err
	}

	return &query, nil
}

func (s *SQLQueryStore) List(page, per int) ([]*Query, error) {
	if page < 1 || per < 1 {
		return nil,
			fmt.Errorf("page and per must be greater than 0 (page %v) (per %v)",
				page, per)
	}

	queries := []*Query{}

	queryStr := `SELECT * FROM querycache_queries ORDER BY created_at OFFSET $1 LIMIT $2`
	if err := s.db.Select(&queries, queryStr, (page-1)*per, per); err != nil {
		return nil, err
	}

	return queries, nil
}
