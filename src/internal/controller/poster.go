package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	servicePkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
	"github.com/opentracing/opentracing-go"
)

type posterService interface {
	Get(ctx context.Context, posterID int) (*model.Poster, error)
	Create(ctx context.Context, poster *model.Poster) (int, error)
	Update(ctx context.Context, poster *model.Poster) error
	Delete(ctx context.Context, posterID int) error
}

type PosterHandler struct {
	service posterService
}

func NewPosterHandler(service posterService) *PosterHandler {
	return &PosterHandler{service: service}
}

// @Summary	Get
// @Description	get poster
// @Tags poster
// @Param token query string true "User auth token"
// @Param id query integer true "PosterId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /poster [get]
func (h *PosterHandler) Get(ctx context.Context, posterID int) ([]byte, error) {
	poster, err := h.service.Get(ctx, posterID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("poster not found", "poster_id", posterID)
		return nil, errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while getting poster", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	posterJSON, err := json.Marshal(poster)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling poster", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return posterJSON, nil
}

// @Summary	Create
// @Description	create poster
// @Tags poster
// @Param input body model.Poster true "Poster body"
// @Param token query string true "User auth token"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /poster [post]
func (h *PosterHandler) Create(ctx context.Context, poster *model.Poster) ([]byte, error) {
	id, err := h.service.Create(ctx, poster)
	if err != nil {
		slog.Error("unexpected error occurred while creating poster", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	idJSON, err := json.Marshal(map[string]int{"id": id})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling id", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return idJSON, nil
}

// @Summary	Update
// @Description	update poster
// @Tags poster
// @Param input body model.Poster true "Poster body"
// @Param token query string true "User auth token"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /poster [put]
func (h *PosterHandler) Update(ctx context.Context, poster *model.Poster) error {
	err := h.service.Update(ctx, poster)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("poster not found", "poster_id", poster.ID)
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while updating poster", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

// @Summary	Delete
// @Description	delete poster
// @Tags poster
// @Param token query string true "User auth token"
// @Param id query integer true "PosterId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /poster [delete]
func (h *PosterHandler) Delete(ctx context.Context, posterID int) error {
	if err := h.service.Delete(ctx, posterID); errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("poster not found", "poster_id", posterID)
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while deleting poster", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

//nolint:funlen,cyclop // http handler methods router
func (c *Controller) handlePosterRequests(w http.ResponseWriter, r *http.Request) {
	span, _ := opentracing.StartSpanFromContext(r.Context(), "app-http-handler: poster")
	defer span.Finish()

	token := r.URL.Query().Get("token")
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		posterID, err := parseInt(r, "id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		poster, err := c.poster.Get(ctx, posterID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(poster); err != nil {
			writeError(w, fmt.Errorf("%w: writing poster body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodPost:
		poster, err := c.parsePoster(r, token)
		if err != nil {
			writeError(w, err)
			break
		}

		id, err := c.poster.Create(ctx, poster)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(id); err != nil {
			writeError(w, fmt.Errorf("%w: writing poster_id body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		poster, err := c.parsePoster(r, token)
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.poster.Update(ctx, poster); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		posterID, err := parseInt(r, "id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.poster.Delete(ctx, posterID); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) authorizePoster(ctx context.Context, token string, posterID int) error {
	userID, err := c.auth.service.GetUserID(token)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("token not found in sessions")
		return fmt.Errorf("%w: token not found in sessions", errNotFound)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting user_id by token", "error", err)
		return err
	}

	poster, err := c.poster.service.Get(ctx, posterID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("poster not found", "poster_id", posterID)
		return fmt.Errorf("%w: poster not found", errNotFound)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting poster", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	if poster.UserID != userID {
		slog.Warn("poster action is not authorized: posters' user_id is different",
			"user_id", userID, "poster_user_id", poster.UserID)
		return errActionNotAuthorized
	}

	return nil
}

func (c *Controller) parsePoster(r *http.Request, token string) (*model.Poster, error) {
	var poster model.Poster
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("poster body cannot be read", "error", err)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	if err = json.Unmarshal(body, &poster); err != nil {
		slog.Warn("poster cannot be unmarshalled", "error", err)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	if err = c.authorizeUserID(token, poster.UserID); err != nil {
		return nil, err
	}

	if err = validatePoster(&poster); err != nil {
		return nil, err
	}

	return &poster, nil
}

func validatePoster(poster *model.Poster) error {
	switch {
	case poster.UserID == 0:
		return fmt.Errorf("%w: poster user_id is absent", errInvalidArguments)
	case poster.Name == "":
		return fmt.Errorf("%w: poster name is absent", errInvalidArguments)
	}

	return nil
}
