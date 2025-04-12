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
	Name     string   `json:"name"`
	Year     int      `json:"year"`
	Genres   []string `json:"genres"`
	Chrono   int      `json:"chrono"`
	KPID     string   `json:"kp_id"`
	ImageURL string   `json:"image_url"`
}

type PosterKPRequest struct {
	KPID string `json:"kp_id"`
}

type PosterResponse struct {
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
		Name:     req.Name,
		Year:     req.Year,
		Genres:   req.Genres,
		UserID:   userID,
		Chrono:   req.Chrono,
		KPID:     req.KPID,
		ImageURL: req.ImageURL,
	}, nil
}

func ParsePosterKPRequest(r *http.Request, userID int) (string, error) {
	var req PosterKPRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", fmt.Errorf("request body cannot be read: %w", err)
	}

	if err = json.Unmarshal(body, &req); err != nil {
		return "", fmt.Errorf("request cannot be unmarshalled: %w", err)
	}

	return req.KPID, nil
}

func ToPosterResponse(poster *svcModel.Poster) *PosterResponse {
	return &PosterResponse{
		ID:        poster.ID,
		Name:      poster.Name,
		Year:      poster.Year,
		Genres:    poster.Genres,
		Chrono:    poster.Chrono,
		UserID:    poster.UserID,
		KPID:      poster.KPID,
		ImageURL:  poster.ImageURL,
		CreatedAt: poster.CreatedAt,
	}
}
