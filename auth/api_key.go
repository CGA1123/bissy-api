package auth

import (
	"fmt"
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

// APIKey represents a key that can be used via to interact with an API as an
// authenticated user.
//
// The Key itself is not exposed.
type APIKey struct {
	ID        string
	Name      string
	UserID    string    `json:"userId" db:"user_id"`
	LastUsed  time.Time `json:"lastUsed" db:"last_used"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// CreateAPIKey defines the parameters passed when creating an APIKey
type CreateAPIKey struct {
	Name string
}

// NewAPIKey represents a newly created API key and is the only struct exposing
// the Key itself
type NewAPIKey struct {
	ID        string
	UserID    string `json:"userId" db:"user_id"`
	Name      string
	Key       string
	LastUsed  time.Time `json:"lastUsed" db:"last_used"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// The APIKeyStore interface defines functions for interacting and managing API
// keys
type APIKeyStore interface {
	Create(string, *CreateAPIKey) (*NewAPIKey, error)
	Delete(string, string) (*APIKey, error)
	GetByKey(string) (*APIKey, error)
	List(string) ([]*APIKey, error)
}

// SQLAPIKeyStore is an SQL-backed implementation of an APIKeyStore
type SQLAPIKeyStore struct {
	db           *hnysqlx.DB
	clock        utils.Clock
	idGenerator  utils.IDGenerator
	keyGenerator utils.Random
}

// NewSQLAPIKeyStore build a new SQLAPIKeyStore
func NewSQLAPIKeyStore(db *hnysqlx.DB) *SQLAPIKeyStore {
	return &SQLAPIKeyStore{
		db:           db,
		clock:        &utils.RealClock{},
		idGenerator:  &utils.UUIDGenerator{},
		keyGenerator: &utils.SecureRandom{},
	}
}

// NewTestSQLAPIKeyStore allow build a SQLAPIKeyStore with custom generators
func NewTestSQLAPIKeyStore(db *hnysqlx.DB, time time.Time, id, key string) *SQLAPIKeyStore {
	return &SQLAPIKeyStore{
		db:           db,
		clock:        &utils.TestClock{Time: time},
		idGenerator:  &utils.TestIDGenerator{ID: id},
		keyGenerator: &utils.TestRandom{Value: []byte(key)},
	}
}

// Create persists a new key for a user
func (store *SQLAPIKeyStore) Create(userID string, ck *CreateAPIKey) (*NewAPIKey, error) {
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

	var newKey NewAPIKey
	if err := store.db.Get(&newKey, query, id, userID, ck.Name, key, now, now); err != nil {
		return nil, err
	}

	return &newKey, nil
}

// Delete removes a give key for a user
func (store *SQLAPIKeyStore) Delete(userID, keyID string) (*APIKey, error) {
	var key APIKey

	query := `
		DELETE FROM auth_api_keys
		WHERE id = $1 AND user_id = $2
		RETURNING id, name, user_id, last_used, created_at`
	if err := store.db.Get(&key, query, keyID, userID); err != nil {
		return nil, err
	}

	return &key, nil
}

// GetByKey returns the APIKey related to a given Key
func (store *SQLAPIKeyStore) GetByKey(key string) (*APIKey, error) {
	var apiKey APIKey

	query := `
		SELECT id, name, user_id, last_used, created_at
		FROM auth_api_keys
		WHERE key = $1`
	if err := store.db.Get(&apiKey, query, key); err != nil {
		return nil, err
	}

	return &apiKey, nil
}

// List returns all APIKeys associated with a user
func (store *SQLAPIKeyStore) List(userID string) ([]*APIKey, error) {
	keys := []*APIKey{}
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
