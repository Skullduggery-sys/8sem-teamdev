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
	"github.com/samber/lo"
)

type listPosterService interface {
	GetPosters(ctx context.Context, listID int) ([]*model.ListPoster, error)
	AddPoster(ctx context.Context, listID, posterID int) error
	MovePoster(ctx context.Context, curListID, newListID, posterID int) error
	ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error
	DeletePoster(ctx context.Context, listID, posterID int) error
}

type ListPosterHandler struct {
	listPosterSvc listPosterService
}

func NewListPosterHandler(service listPosterService) *ListPosterHandler {
	return &ListPosterHandler{listPosterSvc: service}
}

// @Summary	Get posters in list
// @Description	get posters in list
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param list_id path integer true "ListId"
// @Success	200 {array} reqModelPkg.ListPosterResponse
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists/{list_id}/posters [get]
func (h *ListPosterHandler) GetPosters(ctx context.Context, listID int) ([]byte, error) {
	listPosters, err := h.listPosterSvc.GetPosters(ctx, listID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("posters not found by list id", "list_id", listID)
		return nil, fmt.Errorf("%w: %w", errNotFound, err)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting list posters by list id", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	listPostersResponse := lo.Map(listPosters, func(item *model.ListPoster, _ int) *reqModelPkg.ListPosterResponse {
		return reqModelPkg.ToListPosterResponse(item)
	})
	listPostersJSON, err := json.Marshal(listPostersResponse)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling list posters", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return listPostersJSON, nil
}

// @Summary	Add poster in list
// @Description	Adds poster in list. If poster already exists in some list, it will be moved to new list.
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param list_id path integer true "ListId"
// @Param poster_id path integer true "PosterId"
// @Success	200 "Poster moved"
// @Success	201 "Poster added"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists/{list_id}/posters/{poster_id} [post]
func (h *ListPosterHandler) AddPoster(ctx context.Context, listID, posterID int) error {
	err := h.listPosterSvc.AddPoster(ctx, listID, posterID)
	if errors.Is(err, servicePkg.ErrCreated) {
		slog.Debug("added poster", "list_id", listID, "poster_id", posterID)
		return errOkCreated
	} else if err != nil {
		slog.Error("unexpected error occurred while adding poster in list", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	slog.Debug("moved poster", "new_list_id", listID, "poster_id", posterID)
	return nil
}

// @Summary	Change poster position in list
// @Description	change poster position in list
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param list_id path integer true "ListId"
// @Param poster_id path integer true "PosterId"
// @Param position body reqModelPkg.ListPositionRequest true "Change position body"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists/{list_id}/posters/{poster_id} [put]
func (h *ListPosterHandler) ChangePosterPosition(ctx context.Context, listID, posterID, newPosition int) error {
	if err := h.listPosterSvc.ChangePosterPosition(ctx, listID, posterID, newPosition); err != nil {
		slog.Error("unexpected error occurred while changing poster position", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

// @Summary	Delete poster from list
// @Description	delete poster from list
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param list_id path integer true "ListId"
// @Param poster_id path integer true "PosterId"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists/{list_id}/posters/{poster_id} [delete]
func (h *ListPosterHandler) DeletePoster(ctx context.Context, listID, posterID int) error {
	if err := h.listPosterSvc.DeletePoster(ctx, listID, posterID); errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("poster in list not found", "list_id", listID)
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while deleting poster in list", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

func (c *Controller) handleListPosterRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

	ctx := r.Context()
	listID, err := parseInt(r, "list_id")
	if err == nil {
		err = c.authorizeList(ctx, token, listID)
	}
	if err != nil {
		writeError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		listPosters, err := c.listPoster.GetPosters(ctx, listID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(listPosters); err != nil {
			writeError(w, fmt.Errorf("%w: writing list posters body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodPost:
		posterID, err := parseInt(r, "poster_id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		if addErr := c.listPoster.AddPoster(ctx, listID, posterID); errors.Is(addErr, errOkCreated) {
			w.WriteHeader(http.StatusCreated)
			break
		} else if addErr != nil {
			writeError(w, addErr)
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		posterID, err := parseInt(r, "poster_id")
		if err == nil {
			err = c.authorizePoster(ctx, token, posterID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		position, err := reqModelPkg.ParseListPositionRequest(r)
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.listPoster.ChangePosterPosition(ctx, listID, posterID, position); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
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

func (c *Controller) handleGetListPostersRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

	ctx := r.Context()
	listID, err := parseInt(r, "list_id")
	if err == nil {
		err = c.authorizeList(ctx, token, listID)
	}
	if err != nil {
		writeError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		listPosters, err := c.listPoster.GetPosters(ctx, listID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(listPosters); err != nil {
			writeError(w, fmt.Errorf("%w: writing list posters body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}
