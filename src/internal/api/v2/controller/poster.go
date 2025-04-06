package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	reqModelPkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/api/v2/model"
	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	servicePkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
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
// @Tags posters/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param id path integer true "PosterId"
// @Accept json
// @Success	200 {object} reqModelPkg.PosterResponse "Poster"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/posters/{id} [get]
func (h *PosterHandler) Get(ctx context.Context, posterID int) ([]byte, error) {
	poster, err := h.service.Get(ctx, posterID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("poster not found", "poster_id", posterID)
		return nil, fmt.Errorf("%w: %w", errNotFound, err)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting poster", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	posterResponse := reqModelPkg.ToPosterResponse(poster)
	posterJSON, err := json.Marshal(posterResponse)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling poster", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return posterJSON, nil
}

// @Summary	Create
// @Description	create poster
// @Tags posters/v2
// @Param input body reqModelPkg.PosterRequest true "Poster body"
// @Param X-User-Token header string true "JWT-format token"
// @Accept json
// @Success	201 {object} reqModelPkg.IDResponse "id"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/posters [post]
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
// @Tags posters/v2
// @Param input body reqModelPkg.PosterRequest true "Poster body"
// @Param id path integer true "PosterId"
// @Param X-User-Token header string true "JWT-format token"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/posters/{id} [put]
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
// @Tags posters/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param id path integer true "PosterId"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/posters/{id} [delete]
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

func (c *Controller) handlePosterPathRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

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
	case http.MethodPut:
		sess, err := parseJWT(token)
		if err != nil {
			writeError(w, fmt.Errorf("%w: %w", errInternal, err))
			break
		}

		posterID, err := parseInt(r, "id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		poster, err := reqModelPkg.ParsePosterRequest(r, sess.UserID)
		if err != nil {
			writeError(w, err)
			break
		}

		poster.ID = posterID
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

func (c *Controller) handlePosterBodyRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		sess, err := parseJWT(token)
		if err != nil {
			writeError(w, fmt.Errorf("%w: %w", errInternal, err))
			break
		}

		poster, err := reqModelPkg.ParsePosterRequest(r, sess.UserID)
		if err != nil {
			writeError(w, fmt.Errorf("%w: %w", errInvalidArguments, err))
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

		w.WriteHeader(http.StatusCreated)
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

	// TODO add token.role == admin check to skip
	if poster.UserID != userID {
		slog.Warn("poster action is not authorized: posters' user_id is different",
			"user_id", userID, "poster_user_id", poster.UserID)
		return errActionNotAuthorized
	}

	return nil
}
