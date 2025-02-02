package model

type List struct {
	ID       int    `json:"id"`
	ParentID int    `json:"parentId"`
	Name     string `json:"name"`
	UserID   int    `json:"userId"`
}
