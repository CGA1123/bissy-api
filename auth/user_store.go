package auth

import (
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/honeycombio/beeline-go/wrappers/hnysqlx"
)

// User represents an auth user
type User struct {
	ID        string `json:"id"`
	GithubID  string `json:"githubId" db:"github_id"`
	Name      string
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// CreateUser is the struct containing all parameters necessary to create a user
type CreateUser struct {
	GithubID string `json:"github_id"`
	Name     string
}

// UserStore defines the interface to manage users
type UserStore interface {
	Get(string) (*User, error)
	GetByGithubID(string) (*User, error)
	Create(*CreateUser) (*User, error)
}

// NewToken generates a new jwt.Token expiring at exp for the given user
func (u *User) NewToken(exp time.Time) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, Claims{
		u.ID,
		u.Name,
		jwt.StandardClaims{
			ExpiresAt: exp.Unix(),
			Issuer:    "bissy-api",
		},
	})
}

// SQLUserStore is an SQL implementation of UserStore
type SQLUserStore struct {
	db          *hnysqlx.DB
	idGenerator utils.IDGenerator
	clock       utils.Clock
}

// NewSQLUserStore configures a new SQLUserStore
func NewSQLUserStore(db *hnysqlx.DB, clock utils.Clock, gen utils.IDGenerator) *SQLUserStore {
	return &SQLUserStore{db: db, idGenerator: gen, clock: clock}
}

// GetByGithubID fetches a user from the store based on their GithubID
func (s *SQLUserStore) GetByGithubID(id string) (*User, error) {
	var user User

	query := "SELECT * FROM auth_users WHERE github_id = $1"

	if err := s.db.Get(&user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}

// Get fetches a user from the store based on their ID
func (s *SQLUserStore) Get(id string) (*User, error) {
	var user User

	query := "SELECT * FROM auth_users WHERE id = $1"

	if err := s.db.Get(&user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}

// Create persists a new user to the store
func (s *SQLUserStore) Create(cu *CreateUser) (*User, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	query := `
		INSERT INTO auth_users (id, github_id, name, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING *`

	var user User
	if err := s.db.Get(&user, query, id, cu.GithubID, cu.Name, now); err != nil {
		return nil, err
	}

	return &user, nil
}
