package robert

import (
	"encoding/json"
	"errors"
	"time"
)

type Query struct {
	Id          string    `json:"id"`
	Query       string    `json:"query"`
	Lifetime    Duration  `json:"lifetime"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	LastRefresh time.Time `json:"lastRefresh"`
}

type CreateQuery struct {
	Query    string   `json:"query"`
	Lifetime Duration `json:"lifetime"`
}

type UpdateQuery struct {
	Query    *string   `json:"query"`
	Lifetime *Duration `json:"lifetime"`
}

type Duration time.Duration

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

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