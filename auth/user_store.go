package auth

import (
	"time"

	"github.com/cga1123/bissy-api/utils"
	"github.com/dgrijalva/jwt-go"
	"github.com/jmoiron/sqlx"
)

type User struct {
	Id        string `json:"id"`
	GithubId  string `json:"githubId" db:"github_id"`
	Name      string
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type CreateUser struct {
	GithubId string `json:"github_id"`
	Name     string
}

type UserStore interface {
	Get(string) (*User, error)
	GetByGithubId(string) (*User, error)
	Create(*CreateUser) (*User, error)
}

func (u *User) NewToken(exp time.Time) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, BissyClaims{
		u.Id,
		u.Name,
		jwt.StandardClaims{
			ExpiresAt: exp.Unix(),
			Issuer:    "bissy-api",
		},
	})
}

type SQLUserStore struct {
	db          *sqlx.DB
	idGenerator utils.IdGenerator
	clock       utils.Clock
}

func NewSQLUserStore(db *sqlx.DB, clock utils.Clock, gen utils.IdGenerator) *SQLUserStore {
	return &SQLUserStore{db: db, idGenerator: gen, clock: clock}
}

func (s *SQLUserStore) GetByGithubId(id string) (*User, error) {
	var user User

	query := "SELECT * FROM auth_users WHERE github_id = $1"

	if err := s.db.Get(&user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SQLUserStore) Get(id string) (*User, error) {
	var user User

	query := "SELECT * FROM auth_users WHERE id = $1"

	if err := s.db.Get(&user, query, id); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *SQLUserStore) Create(cu *CreateUser) (*User, error) {
	now := s.clock.Now()
	id := s.idGenerator.Generate()

	query := `
		INSERT INTO auth_users (id, github_id, name, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING *`

	var user User
	if err := s.db.Get(&user, query, id, cu.GithubId, cu.Name, now); err != nil {
		return nil, err
	}

	return &user, nil
}
