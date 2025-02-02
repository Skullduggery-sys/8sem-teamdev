package model

import "time"

type PosterRecord struct {
	ID        int       `json:"id"`
	PosterID  int       `json:"posterId"`
	UserID    int       `json:"userId"`
	CreatedAt time.Time `json:"createdat"`
}
