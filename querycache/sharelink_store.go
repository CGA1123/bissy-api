package querycache

import "time"

// A Sharelink is used to expose the results of a specific query to the public
// A Query may have many sharelinks, which can be individually managed
type Sharelink struct {
	ID        string    `json:"id" db:"id"`
	QueryID   string    `json:"queryId" db:"query_id"`
	UserID    string    `json:"userId" db:"user_id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

// The SharelinkStore is used to manage the collection of Sharelinks
type SharelinkStore interface {
	Create(*Query) (*Sharelink, error)
	ListForQuery(*Query) ([]*Sharelink, error)
	ListForUser(string) ([]*Sharelink, error)
	Get(string) (*Sharelink, error)
}
