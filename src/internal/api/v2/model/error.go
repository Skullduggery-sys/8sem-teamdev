package model

type ErrorResponse struct {
	Error string `json:"error" example:"error description"`
}

type IDResponse struct {
	ID int `json:"id"`
}
