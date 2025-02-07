package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	servicePkg "git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
)

type listPosterService interface {
	GetPosters(ctx context.Context, listID int) ([]*model.ListPoster, error)
	AddPoster(ctx context.Context, listID, posterID int) error
	MovePoster(ctx context.Context, curListID, newListID, posterID int) error
	ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error
	DeletePoster(ctx context.Context, listID, posterID int) error
}

type ListPosterHandler struct {
	service listPosterService
}

func NewListPosterHandler(service listPosterService) *ListPosterHandler {
	return &ListPosterHandler{service: service}
}

// @Summary	Get posters in list
// @Description	get posters in list
// @Tags list-poster/v1
// @Param token query string true "User auth token"
// @Param list_id query integer true "ListId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /api/v1/list-poster [get]
func (h *ListPosterHandler) GetPosters(ctx context.Context, listID int) ([]byte, error) {
	listPosters, err := h.service.GetPosters(ctx, listID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("posters not found by list id", "list_id", listID)
		return nil, errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while getting list posters by list id", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	listPostersJSON, err := json.Marshal(listPosters)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling list posters", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return listPostersJSON, nil
}

// @Summary	Add poster in list
// @Description	add poster in list
// @Tags list-poster/v1
// @Param token query string true "User auth token"
// @Param list_id query integer true "ListId"
// @Param poster_id query integer true "PosterId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /api/v1/list-poster/add [post]
func (h *ListPosterHandler) AddPoster(ctx context.Context, listID, posterID int) error {
	if err := h.service.AddPoster(ctx, listID, posterID); err != nil {
		slog.Error("unexpected error occurred while adding poster in list", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

// @Summary	Move poster from one list to another
// @Description	move poster from one list to another
// @Tags list-poster/v1
// @Param token query string true "User auth token"
// @Param cur_list_id query integer true "CurListId"
// @Param new_list_id query integer true "NewListId"
// @Param poster_id query integer true "PosterId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /api/v1/list-poster/move [post]
func (h *ListPosterHandler) MovePoster(ctx context.Context, curListID, newListID, posterID int) error {
	if err := h.service.MovePoster(ctx, curListID, newListID, posterID); err != nil {
		slog.Error("unexpected error occurred while moving poster", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

// @Summary	Change poster position in list
// @Description	change poster position in list
// @Tags list-poster/v1
// @Param token query string true "User auth token"
// @Param list_id query integer true "ListId"
// @Param poster_id query integer true "PosterId"
// @Param position query integer true "NewPosition"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /api/v1/list-poster/change-position [post]
func (h *ListPosterHandler) ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error {
	if err := h.service.ChangePosterPosition(ctx, listID, posterID, newPosition); err != nil {
		slog.Error("unexpected error occurred while changing poster position", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

// @Summary	Delete poster from list
// @Description	delete poster from list
// @Tags list-poster/v1
// @Param token query string true "User auth token"
// @Param list_id query integer true "ListId"
// @Param poster_id query integer true "PosterId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /api/v1/list-poster [delete]
func (h *ListPosterHandler) DeletePoster(ctx context.Context, listID, posterID int) error {
	if err := h.service.DeletePoster(ctx, listID, posterID); errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("poster in list not found", "list_id", listID)
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while deleting poster in list", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

func (c *Controller) handleListPosterGetDeleteRequests(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		listID, err := parseInt(r, "list_id")
		if err == nil {
			err = c.authorizeList(ctx, token, listID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		listPosters, err := c.listPoster.GetPosters(ctx, listID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(listPosters); err != nil {
			writeError(w, fmt.Errorf("%w: writing list posters body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		listID, err := parseInt(r, "list_id")
		if err == nil {
			err = c.authorizeList(ctx, token, listID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		posterID, err := parseInt(r, "poster_id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.listPoster.DeletePoster(ctx, listID, posterID); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) handleListPosterAddRequests(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		listID, err := parseInt(r, "list_id")
		if err == nil {
			err = c.authorizeList(ctx, token, listID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		posterID, err := parseInt(r, "poster_id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.listPoster.AddPoster(ctx, listID, posterID); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) handleListPosterMoveRequests(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		curListID, err := parseInt(r, "cur_list_id")
		if err == nil {
			err = c.authorizeList(ctx, token, curListID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		newListID, err := parseInt(r, "new_list_id")
		if err == nil {
			err = c.authorizeList(ctx, token, newListID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		posterID, err := parseInt(r, "poster_id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.listPoster.MovePoster(ctx, curListID, newListID, posterID); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) handleListPosterChangePositionRequests(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		listID, err := parseInt(r, "list_id")
		if err == nil {
			err = c.authorizeList(ctx, token, listID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		posterID, err := parseInt(r, "poster_id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		position, err := parseInt(r, "position")
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.listPoster.ChangePosterPosition(ctx, listID, posterID, position); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}
