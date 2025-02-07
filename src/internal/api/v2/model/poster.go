package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	svcModel "git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
)

type PosterRequest struct {
	Name   string   `json:"name"`
	Year   int      `json:"year"`
	Genres []string `json:"genres"`
	Chrono int      `json:"chrono"`
}

type PosterResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Year      int       `json:"year"`
	Genres    []string  `json:"genres"`
	Chrono    int       `json:"chrono"`
	UserID    int       `json:"userId"`
	CreatedAt time.Time `json:"createdat"` // will not be used, satisfy musttag linter
}

func ParsePosterRequest(r *http.Request, userID int) (*svcModel.Poster, error) {
	var req PosterRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("request body cannot be read: %w", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("request cannot be unmarshalled: %w", err)
	}

	return &svcModel.Poster{
		Name:   req.Name,
		Year:   req.Year,
		Genres: req.Genres,
		UserID: userID,
		Chrono: req.Chrono,
	}, nil
}

func ToPosterResponse(poster *svcModel.Poster) *PosterResponse {
	return &PosterResponse{
		ID:        poster.ID,
		Name:      poster.Name,
		Year:      poster.Year,
		Genres:    poster.Genres,
		Chrono:    poster.Chrono,
		UserID:    poster.UserID,
		CreatedAt: poster.CreatedAt,
	}
}
