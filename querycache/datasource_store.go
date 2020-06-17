package querycache

import (
	"time"
)

// Datasource describes a database that Queries may be related to and executed
// against
type Datasource struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Options   string    `json:"options"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// UpdateDatasource describes the paramater which may be updated on a Datasource
type UpdateDatasource struct {
	Name    *string `json:"name"`
	Type    *string `json:"type"`
	Options *string `json:"options"`
}

// CreateDatasource describes the required paramater to create a new Datasource
type CreateDatasource struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Options string `json:"options"`
}

// DatasourceStore describes a generic Store for Datasources
type DatasourceStore interface {
	Get(string) (*Datasource, error)
	Create(string, *CreateDatasource) (*Datasource, error)
	List(page int, per int) ([]*Datasource, error)
	Delete(string) (*Datasource, error)
	Update(string, *UpdateDatasource) (*Datasource, error)
}

// NewExecutor returns a new SQLExecutor configured against this Datasource
// Will return a TestExecutor if the datasources "Type" is "test"
func (a *Datasource) NewExecutor() (Executor, error) {
	switch a.Type {
	case "test":
		return &TestExecutor{}, nil
	default:
		// TODO: Cache this per datasource-id? keep DB objects available and not need to recreate connections?
		return NewSQLExecutor(a.Type, a.Options)
	}
}
