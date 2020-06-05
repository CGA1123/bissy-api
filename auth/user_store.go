package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type IdGenerator interface {
	Generate() string
}

type UUIDGenerator struct{}

func (generator *UUIDGenerator) Generate() string {
	return uuid.New().String()
}

type User struct {
	Id        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type CreateUser struct {
	Email string
}

type UserStore interface {
	Get(string) (*User, error)
	Create(*CreateUser) (*User, error)
}

func (u *User) NewToken(exp time.Time) *jwt.Token {
	return jwt.NewWithClaims(jwt.SigningMethodHS512, BissyClaims{
		u.Id,
		jwt.StandardClaims{
			ExpiresAt: exp.Unix(),
			Issuer:    "bissy-api",
		},
	})
}

type SQLUserStore struct {
	db          *sqlx.DB
	idGenerator IdGenerator
	clock       Clock
}

func NewSQLUserStore(db *sqlx.DB, clock Clock, gen IdGenerator) *SQLUserStore {
	return &SQLUserStore{db: db, idGenerator: gen, clock: clock}
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
		INSERT INTO auth_users (id, email, created_at)
		VALUES ($1, $2, $3)
		RETURNING *`

	var user User
	if err := s.db.Get(&user, query, id, cu.Email, now); err != nil {
		return nil, err
	}

	return &user, nil
}
