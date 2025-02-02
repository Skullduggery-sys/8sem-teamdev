package model

type ListPoster struct {
	ID       int `json:"id"`
	ListID   int `json:"listId"`
	PosterID int `json:"posterId"`
	Position int `json:"position"`
}
