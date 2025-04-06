package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	svcModel "git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
)

type ListCreateRequest struct {
	ParentID int    `json:"parentId"`
	Name     string `json:"name"`
}

type ListUpdateRequest struct {
	Name string `json:"name"`
}

type ListResponse struct {
	ID       int    `json:"id"`
	ParentID int    `json:"parentId"`
	Name     string `json:"name"`
	UserID   int    `json:"userId"`
}

type ListPosterResponse struct {
	// ID       int `json:"id"`
	ListID   int `json:"listId"`
	PosterID int `json:"posterId"`
	Position int `json:"position"`
}

type ListPositionRequest struct {
	Position int `json:"position"`
}

func ParseListCreateRequest(r *http.Request, userID int) (*svcModel.List, error) {
	var req ListCreateRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("request body cannot be read: %w", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("request cannot be unmarshalled: %w", err)
	}

	return &svcModel.List{
		Name:     req.Name,
		ParentID: req.ParentID,
		UserID:   userID,
	}, nil
}

func ParseListUpdateRequest(r *http.Request, userID int) (*svcModel.List, error) {
	var req ListCreateRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("request body cannot be read: %w", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("request cannot be unmarshalled: %w", err)
	}

	return &svcModel.List{
		Name:     req.Name,
		ParentID: req.ParentID,
		UserID:   userID,
	}, nil
}

func ParseListPositionRequest(r *http.Request) (int, error) {
	var req ListPositionRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return 0, fmt.Errorf("request body cannot be read: %w", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return 0, fmt.Errorf("request cannot be unmarshalled: %w", err)
	}

	return req.Position, nil
}

func ToListResponse(list *svcModel.List) *ListResponse {
	return &ListResponse{
		ID:       list.ID,
		Name:     list.Name,
		ParentID: list.ParentID,
		UserID:   list.UserID,
	}
}

func ToListPosterResponse(listPoster *svcModel.ListPoster) *ListPosterResponse {
	return &ListPosterResponse{
		ListID:   listPoster.ListID,
		PosterID: listPoster.PosterID,
		Position: listPoster.Position,
	}
}
