package querycache

import (
	"encoding/json"
	"errors"
	"time"
)

// Query describes an SQL query on a given datasource that should be cached for
// a given Lifetime value
type Query struct {
	ID           string    `json:"id"`
	UserID       string    `json:"userId" db:"user_id"`
	Query        string    `json:"query"`
	DatasourceID string    `json:"datasourceId" db:"datasource_id"`
	Lifetime     Duration  `json:"lifetime"`
	CreatedAt    time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time `json:"updatedAt" db:"updated_at"`
	LastRefresh  time.Time `json:"lastRefresh" db:"last_refresh"`
}

// Fresh determines whether a query was last refreshed within Lifetime of the
// given time parameter
func (query *Query) Fresh(now time.Time) bool {
	timeSinceLastRefresh := now.Sub(query.LastRefresh)
	refreshedRecently := Duration(timeSinceLastRefresh) < query.Lifetime

	return refreshedRecently
}

// CreateQuery describes the required parameter to create a new Query
type CreateQuery struct {
	Query        string   `json:"query"`
	Lifetime     Duration `json:"lifetime"`
	DatasourceID string   `json:"datasourceId"`
}

// UpdateQuery describes the paramater which may be updated on a Query
type UpdateQuery struct {
	Lifetime    *Duration `json:"lifetime"`
	LastRefresh time.Time `json:"lastRefresh"`
}

// Duration is an alias to time.Duration to allow for defining JSON marshalling
// and unmarshalling
type Duration time.Duration

// MarshalJSON marshals a Duration into JSON
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON unmarshals JSON into a Duration
func (d *Duration) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		*d = Duration(time.Duration(value))
		return nil
	case string:
		tmp, err := time.ParseDuration(value)
		if err != nil {
			return err
		}
		*d = Duration(tmp)
		return nil
	default:
		return errors.New("invalid duration")
	}
}

// QueryStore describes a generic Store for Queries
type QueryStore interface {
	Get(string, string) (*Query, error)
	Create(string, *CreateQuery) (*Query, error)
	List(string, int, int) ([]*Query, error)
	Delete(string, string) (*Query, error)
	Update(string, string, *UpdateQuery) (*Query, error)
}
