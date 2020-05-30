package querycache

import "time"

type Adapter struct {
	Id        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Options   string    `json:"options"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateAdapter struct {
	Name    *string `json:"name"`
	Type    *string `json:"type"`
	Options *string `json:"options"`
}

type CreateAdapter struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Options string `json:"options"`
}
