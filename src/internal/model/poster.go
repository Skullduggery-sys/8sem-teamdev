package model

import "time"

type Poster struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Year      int       `json:"year"`
	Genres    []string  `json:"genres"`
	Chrono    int       `json:"chrono"`
	UserID    int       `json:"userId"`
	CreatedAt time.Time `json:"createdat"` // will not be used, satisfy musttag linter
}
