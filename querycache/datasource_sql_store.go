package querycache

import (
	"fmt"

	"github.com/cga1123/bissy-api/utils"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

// SQLDatasourceStore describes an SQL implementation of DatasourceStore
type SQLDatasourceStore struct {
	db          *hnysqlx.DB
	clock       utils.Clock
	idGenerator utils.IDGenerator
}

// NewSQLDatasourceStore retunes a new SQLDatasourceStore
func NewSQLDatasourceStore(db *hnysqlx.DB, clock utils.Clock, generator utils.IDGenerator) *SQLDatasourceStore {
	return &SQLDatasourceStore{db: db, clock: clock, idGenerator: generator}
}

// Create creates and persists a new Datasource to the Store
func (s *SQLDatasourceStore) Create(userID string, ca *CreateDatasource) (*Datasource, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	query := `
		INSERT INTO querycache_datasources (id, user_id, name, type, options, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING *`

	var datasource Datasource
	if err := s.db.Get(&datasource, query, id, userID, ca.Name, ca.Type, ca.Options, now, now); err != nil {
		return nil, err
	}

	return &datasource, nil
}

// Get returns the Datasource with associated id from the store
func (s *SQLDatasourceStore) Get(userID, id string) (*Datasource, error) {
	var datasource Datasource

	query := "SELECT * FROM querycache_datasources WHERE id = $1 AND user_id = $2"

	if err := s.db.Get(&datasource, query, id, userID); err != nil {
		return nil, err
	}

	return &datasource, nil
}

// Delete removes the Datasource with given id from the Store
func (s *SQLDatasourceStore) Delete(userID, id string) (*Datasource, error) {
	var datasource Datasource

	query := "DELETE FROM querycache_datasources WHERE id = $1 AND user_id = $2 RETURNING *"

	if err := s.db.Get(&datasource, query, id, userID); err != nil {
		return nil, err
	}

	return &datasource, nil
}

// List returns the requests Datasources from the Store, ordered by createdAt
func (s *SQLDatasourceStore) List(page, per int) ([]*Datasource, error) {
	if page < 1 || per < 1 {
		return nil,
			fmt.Errorf("page and per must be greater than 0 (page %v) (per %v)",
				page, per)
	}

	datasources := []*Datasource{}

	query := `SELECT * FROM querycache_datasources ORDER BY created_at OFFSET $1 LIMIT $2`
	if err := s.db.Select(&datasources, query, (page-1)*per, per); err != nil {
		return nil, err
	}

	return datasources, nil
}

// Update updates the Datasource with associated id from the store
func (s *SQLDatasourceStore) Update(userID, id string, ua *UpdateDatasource) (*Datasource, error) {
	var datasource Datasource

	query := `
		UPDATE querycache_datasources
		SET name = COALESCE($3, name),
				type = COALESCE($4, type),
				options = COALESCE($5, options)
		WHERE 1=1
		AND id = $1
		AND user_id = $2
		RETURNING *`

	if err := s.db.Get(&datasource, query, id, userID, ua.Name, ua.Type, ua.Options); err != nil {
		return nil, err
	}

	return &datasource, nil
}
