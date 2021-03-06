package querycache

import (
	"database/sql"
	"fmt"

	"github.com/cga1123/bissy-api/utils"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

// SQLQueryStore defines an SQL implementation of a QueryStore
type SQLQueryStore struct {
	db          *hnysqlx.DB
	clock       utils.Clock
	idGenerator utils.IDGenerator
}

// NewSQLQueryStore builds a new SQLQueryStore
func NewSQLQueryStore(db *hnysqlx.DB, clock utils.Clock, generator utils.IDGenerator) *SQLQueryStore {
	return &SQLQueryStore{db: db, clock: clock, idGenerator: generator}
}

// Get returns the Query with associated id from the store
func (s *SQLQueryStore) Get(userID, id string) (*Query, error) {
	var query Query

	queryStr := "SELECT * FROM querycache_queries WHERE id = $1 AND user_id = $2"

	if err := s.db.Get(&query, queryStr, id, userID); err != nil {
		return nil, err
	}

	return &query, nil
}

// Create creates and persist to memory a Query from a CreateQuery struct
func (s *SQLQueryStore) Create(userID string, ca *CreateQuery) (*Query, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	queryStr := `
		INSERT INTO querycache_queries (id, user_id, query, lifetime, datasource_id, created_at, updated_at, last_refresh)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING *`

	var query Query
	if err := s.db.Get(&query, queryStr, id, userID, ca.Query, ca.Lifetime, ca.DatasourceID, now, now, now); err != nil {
		return nil, err
	}

	return &query, nil
}

// Delete removes the Query with associated id from the store
func (s *SQLQueryStore) Delete(userID, id string) (*Query, error) {
	var query Query

	queryStr := "DELETE FROM querycache_queries WHERE id = $1 AND user_id = $2 RETURNING *"

	if err := s.db.Get(&query, queryStr, id, userID); err != nil {
		return nil, err
	}

	return &query, nil
}

// Update updates the Query with associated id from the store
func (s *SQLQueryStore) Update(userID, id string, uq *UpdateQuery) (*Query, error) {
	var query Query

	queryStr := `
		UPDATE querycache_queries
		SET lifetime = COALESCE($3, lifetime),
				last_refresh = COALESCE($4, last_refresh),
				updated_at = $5
		WHERE 1=1
		AND id = $1
		AND user_id = $2
		RETURNING *`

	var lastRefresh sql.NullTime
	if uq.LastRefresh.IsZero() {
		lastRefresh = sql.NullTime{Valid: false}
	} else {
		lastRefresh = sql.NullTime{Time: uq.LastRefresh, Valid: true}
	}

	err := s.db.Get(&query, queryStr, id, userID, uq.Lifetime, lastRefresh, s.clock.Now())
	if err != nil {
		return nil, err
	}

	return &query, nil
}

// List returns the requests Queries from the Store, ordered by createdAt
func (s *SQLQueryStore) List(userID string, page, per int) ([]*Query, error) {
	if page < 1 || per < 1 {
		return nil,
			fmt.Errorf("page and per must be greater than 0 (page %v) (per %v)",
				page, per)
	}

	queries := []*Query{}

	queryStr := `
		SELECT *
		FROM querycache_queries
		WHERE user_id = $1
		ORDER BY created_at
		OFFSET $2
		LIMIT $3`
	if err := s.db.Select(&queries, queryStr, userID, (page-1)*per, per); err != nil {
		return nil, err
	}

	return queries, nil
}
