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

type listService interface {
	Get(ctx context.Context, listID int) (*model.List, error)
	GetSublists(ctx context.Context, listID int) ([]*model.List, error)
	GetUserRoot(ctx context.Context, userID int) (*model.List, error)
	Create(ctx context.Context, list *model.List) (int, error)
	Update(ctx context.Context, list *model.List) error
	Delete(ctx context.Context, listID int) error
}

type ListHandler struct {
	service listService
}

func NewListHandler(service listService) *ListHandler {
	return &ListHandler{service: service}
}

// @Summary	Get
// @Description	get list
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param id path integer true "ListId"
// @Accept json
// @Success	200 {object} reqModelPkg.ListResponse
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists/{id} [get]
func (h *ListHandler) Get(ctx context.Context, listID int) ([]byte, error) {
	list, err := h.service.Get(ctx, listID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("list not found", "list_id", listID)
		return nil, fmt.Errorf("%w: %w", errNotFound, err)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	listResponse := reqModelPkg.ToListResponse(list)
	listJSON, err := json.Marshal(listResponse)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return listJSON, nil
}

// @Summary	Get user root list
// @Description	get user root list
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Accept json
// @Success	200 {object} reqModelPkg.ListResponse
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists-root [get]
func (h *ListHandler) GetUserRoot(ctx context.Context, userID int) ([]byte, error) {
	list, err := h.service.GetUserRoot(ctx, userID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("list not found", "list_id", userID)
		return nil, fmt.Errorf("%w: %w", errNotFound, err)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	listResponse := reqModelPkg.ToListResponse(list)
	listJSON, err := json.Marshal(listResponse)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return listJSON, nil
}

// @Summary	Get sublists
// @Description	get sublists of the list
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param id path integer true "ListId"
// @Accept json
// @Success	200 {array} reqModelPkg.ListResponse
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/sublists/{id} [get]
func (h *ListHandler) GetSublists(ctx context.Context, listID int) ([]byte, error) {
	lists, err := h.service.GetSublists(ctx, listID)
	if err != nil {
		slog.Error("unexpected error occurred while getting list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	listsResponse := lo.Map(lists, func(list *model.List, _ int) *reqModelPkg.ListResponse {
		return &reqModelPkg.ListResponse{
			ID:       list.ID,
			ParentID: list.ParentID,
			Name:     list.Name,
			UserID:   list.UserID,
		}
	})
	listJSON, err := json.Marshal(listsResponse)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return listJSON, nil
}

// @Summary	Create
// @Description	create list
// @Tags lists/v2
// @Param input body reqModelPkg.ListCreateRequest true "List body"
// @Param X-User-Token header string true "JWT-format token"
// @Accept json
// @Success	201 {object} reqModelPkg.IDResponse "id"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists [post]
func (h *ListHandler) Create(ctx context.Context, list *model.List) ([]byte, error) {
	id, err := h.service.Create(ctx, list)
	if err != nil {
		slog.Error("unexpected error occurred while creating list", "error", err)
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
// @Description	update list
// @Tags lists/v2
// @Param input body reqModelPkg.ListUpdateRequest true "List body"
// @Param id path integer true "ListId"
// @Param X-User-Token header string true "JWT-format token"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists/{id} [put]
func (h *ListHandler) Update(ctx context.Context, list *model.List) error {
	err := h.service.Update(ctx, list)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("list not found", "list_id", list.ID)
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while updating list", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

// @Summary	Delete
// @Description	delete list
// @Tags lists/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param id path integer true "ListId"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/lists/{id} [delete]
func (h *ListHandler) Delete(ctx context.Context, listID int) error {
	if err := h.service.Delete(ctx, listID); errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("list not found", "list_id", listID)
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while deleting list", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

func (c *Controller) handleSublistsRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		listID, err := parseInt(r, "id")
		if err != nil {
			writeError(w, err)
			break
		}

		list, err := c.list.GetSublists(ctx, listID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(list); err != nil {
			writeError(w, fmt.Errorf("%w: writing list body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	}
}

func (c *Controller) handleListUserRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		userID, err := c.getUserIDByToken(token)
		if err != nil {
			writeError(w, err)
			break
		}

		list, err := c.list.GetUserRoot(ctx, userID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(list); err != nil {
			writeError(w, fmt.Errorf("%w: writing list body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	}
}

//nolint:funlen,cyclop // http handler methods router
func (c *Controller) handleListPathRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		listID, err := parseInt(r, "id")
		if err == nil {
			err = c.authorizeList(ctx, token, listID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		list, err := c.list.Get(ctx, listID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(list); err != nil {
			writeError(w, fmt.Errorf("%w: writing list body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		sess, err := parseJWT(token)
		if err != nil {
			writeError(w, fmt.Errorf("%w: %w", errInternal, err))
			break
		}

		listID, err := parseInt(r, "id")
		if err == nil {
			err = c.authorizeList(ctx, token, listID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		list, err := reqModelPkg.ParseListUpdateRequest(r, sess.UserID)
		if err != nil {
			writeError(w, fmt.Errorf("%w: %w", errInvalidArguments, err))
			break
		}

		list.ID = listID
		if err := c.list.Update(ctx, list); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		listID, err := parseInt(r, "id")
		if err == nil {
			err = c.authorizeList(ctx, token, listID)
		}
		if err != nil {
			writeError(w, err)
			break
		}

		if err := c.list.Delete(ctx, listID); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) handleListDefaultRequests(w http.ResponseWriter, r *http.Request) {
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

		list, err := reqModelPkg.ParseListCreateRequest(r, sess.UserID)
		if err != nil {
			writeError(w, fmt.Errorf("%w: %w", errInvalidArguments, err))
			break
		}

		id, err := c.list.Create(ctx, list)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(id); err != nil {
			writeError(w, fmt.Errorf("%w: writing list id body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusCreated)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}

func (c *Controller) authorizeList(ctx context.Context, token string, listID int) error {
	userID, err := c.getUserIDByToken(token)
	if err != nil {
		return err
	}

	list, err := c.list.service.Get(ctx, listID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("list not found", "list_id", listID)
		return fmt.Errorf("%w: list not found", errNotFound)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting list", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	// TODO add token.role == admin check to skip
	if list.UserID != userID {
		slog.Warn("list action is not authorized: lists' user_id is different",
			"user_id", userID, "list_user_id", list.UserID)
		return errActionNotAuthorized
	}

	return nil
}
