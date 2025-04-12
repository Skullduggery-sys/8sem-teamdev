package model

import "time"

type Poster struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Year      int       `json:"year"`
	Genres    []string  `json:"genres"`
	Chrono    int       `json:"chrono"`
	UserID    int       `json:"userId"`
	KPID      string    `json:"kp_id"`
	ImageURL  string    `json:"image_url"`
	CreatedAt time.Time `json:"createdat"` // will not be used, satisfy musttag linter
}
