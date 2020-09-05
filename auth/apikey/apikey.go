package apikey

import (
	"fmt"
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

// Struct represents a key that can be used via to interact with an API as an
// authenticated user.
//
// The Key itself is not exposed.
type Struct struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UserID    string    `json:"userId" db:"user_id"`
	LastUsed  time.Time `json:"lastUsed" db:"last_used"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// Create defines the parameters passed when creating an
type Create struct {
	Name string
}

// New represents a newly created API key and is the only struct exposing
// the Key itself
type New struct {
	ID        string
	UserID    string `json:"userId" db:"user_id"`
	Name      string
	Key       string
	LastUsed  time.Time `json:"lastUsed" db:"last_used"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// The Store interface defines functions for interacting and managing API
// keys
type Store interface {
	Create(string, *Create) (*New, error)
	Delete(string, string) (*Struct, error)
	GetByKey(string) (*Struct, error)
	List(string) ([]*Struct, error)
}

// SQLStore is an SQL-backed implementation of a Store
type SQLStore struct {
	db           *hnysqlx.DB
	clock        utils.Clock
	idGenerator  utils.IDGenerator
	keyGenerator utils.Random
}

// NewSQLStore build a new Store
func NewSQLStore(db *hnysqlx.DB) *SQLStore {
	return &SQLStore{
		db:           db,
		clock:        &utils.RealClock{},
		idGenerator:  &utils.UUIDGenerator{},
		keyGenerator: &utils.SecureRandom{},
	}
}

// NewTestSQLStore allow build a Store with custom generators
func NewTestSQLStore(db *hnysqlx.DB, time time.Time, id, key string) *SQLStore {
	return &SQLStore{
		db:           db,
		clock:        &utils.TestClock{Time: time},
		idGenerator:  &utils.TestIDGenerator{ID: id},
		keyGenerator: &utils.TestRandom{Value: []byte(key)},
	}
}

// Create persists a new key for a user
func (store *SQLStore) Create(userID string, ck *Create) (*New, error) {
	now := store.clock.Now()
	id := store.idGenerator.Generate()
	key, err := store.keyGenerator.String(32)
	if err != nil {
		return nil, fmt.Errorf("error generating key: %v", err)
	}

	query := `
		INSERT INTO auth_api_keys (id, user_id, name, key, last_used, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING *`

	var newKey New
	if err := store.db.Get(&newKey, query, id, userID, ck.Name, key, now, now); err != nil {
		return nil, err
	}

	return &newKey, nil
}

// Delete removes a give key for a user
func (store *SQLStore) Delete(userID, keyID string) (*Struct, error) {
	var key Struct

	query := `
		DELETE FROM auth_api_keys
		WHERE id = $1 AND user_id = $2
		RETURNING id, name, user_id, last_used, created_at`
	if err := store.db.Get(&key, query, keyID, userID); err != nil {
		return nil, err
	}

	return &key, nil
}

// GetByKey returns the  related to a given Key
func (store *SQLStore) GetByKey(key string) (*Struct, error) {
	var apiKey Struct

	query := `
		SELECT id, name, user_id, last_used, created_at
		FROM auth_api_keys
		WHERE key = $1`
	if err := store.db.Get(&apiKey, query, key); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// List returns all s associated with a user
func (store *SQLStore) List(userID string) ([]*Struct, error) {
	keys := []*Struct{}
	query := `
		SELECT id, name, user_id, last_used, created_at
		FROM auth_api_keys
		WHERE user_id = $1
		ORDER BY name
	`

	if err := store.db.Select(&keys, query, userID); err != nil {
		return nil, err
	}

	return keys, nil
}
