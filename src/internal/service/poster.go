package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	repository "git.iu7.bmstu.ru/vai20u117/testing/src/internal/repository/postgres"
	"github.com/samber/lo"
)

type posterRepository interface {
	Get(ctx context.Context, posterID int) (*model.Poster, error)
	Create(ctx context.Context, poster *model.Poster) (int, error)
	Update(ctx context.Context, poster *model.Poster) error
	Delete(ctx context.Context, posterID int) error
}

type PosterService struct {
	repo posterRepository

	kpToken    string
	httpClient *http.Client
}

func NewPosterService(repo posterRepository, httpClient *http.Client, kpToken string) *PosterService {
	return &PosterService{
		repo:       repo,
		httpClient: httpClient,
		kpToken:    kpToken,
	}
}

func (s *PosterService) Get(ctx context.Context, posterID int) (*model.Poster, error) {
	poster, err := s.repo.Get(ctx, posterID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	return poster, nil
}

func (s *PosterService) CreateFromKP(ctx context.Context, kpID string, userID int) (int, error) {
	url := "https://api.kinopoisk.dev/v1.4/movie/" + kpID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("%w: creating GET movie request: %w", ErrKPRequest, err)
	}
	req.Header.Add("X-API-KEY", s.kpToken)
	req.Header.Add("accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%w: got response error on GET movie request: %w", ErrKPRequest, err)
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("%w: failed to read response body of GET movie request: %w", ErrKPRequest, err)
	}

	var movie movieResponse
	if err = json.Unmarshal(rawBody, &movie); err != nil {
		return 0, fmt.Errorf("%w: failed to parse response body of GET movie request: %w", ErrKPRequest, err)
	}

	genres := lo.Map(movie.Genres, func(genre genreResponse, _ int) string { return genre.Name })
	poster := &model.Poster{
		KPID:     kpID,
		Name:     movie.Name,
		Year:     movie.Year,
		Genres:   genres,
		Chrono:   getMovieChrono(movie),
		UserID:   userID,
		ImageURL: movie.Poster.URL,
	}

	slog.Debug("parsed poster from KP", "poster", poster)
	return s.repo.Create(ctx, poster)
}

func (s *PosterService) Create(ctx context.Context, poster *model.Poster) (int, error) {
	return s.repo.Create(ctx, poster)
}

func (s *PosterService) Update(ctx context.Context, poster *model.Poster) error {
	err := s.repo.Update(ctx, poster)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func (s *PosterService) Delete(ctx context.Context, posterID int) error {
	err := s.repo.Delete(ctx, posterID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	} else if err != nil {
		return err
	}

	return nil
}

func getMovieChrono(movie movieResponse) int {
	if !movie.IsSeries {
		return movie.MovieLength
	}

	episodesCount := 0
	for _, item := range movie.SeasonsInfo {
		episodesCount += item.EpisodesCount
	}

	return movie.SeriesLength * episodesCount
}
