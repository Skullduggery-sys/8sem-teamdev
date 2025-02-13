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

type listService interface {
	Get(ctx context.Context, listID int) (*model.List, error)
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
// @Tags list
// @Param token query string true "User auth token"
// @Param id query integer true "ListId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /list [get]
func (h *ListHandler) Get(ctx context.Context, listID int) ([]byte, error) {
	list, err := h.service.Get(ctx, listID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("list not found", "list_id", listID)
		return nil, errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while getting list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	listJSON, err := json.Marshal(list)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling list", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return listJSON, nil
}

// @Summary	Create
// @Description	create list
// @Tags list
// @Param input body model.List true "List body"
// @Param token query string true "User auth token"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /list [post]
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
// @Tags list
// @Param input body model.List true "List body"
// @Param token query string true "User auth token"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /list [put]
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
// @Tags list
// @Param token query string true "User auth token"
// @Param id query integer true "ListId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /list [delete]
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

//nolint:funlen,cyclop // http handler methods router
func (c *Controller) handleListRequests(w http.ResponseWriter, r *http.Request) {
	span, ctx := opentracing.StartSpanFromContext(r.Context(), "app-http-handler: list")
	defer span.Finish()

	token := r.URL.Query().Get("token")

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
	case http.MethodPost:
		list, err := c.parseList(r, token)
		if err != nil {
			writeError(w, err)
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

		w.WriteHeader(http.StatusOK)
	case http.MethodPut:
		list, err := c.parseList(r, token)
		if err != nil {
			writeError(w, err)
			break
		}

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

	if list.UserID != userID {
		slog.Warn("list action is not authorized: lists' user_id is different",
			"user_id", userID, "list_user_id", list.UserID)
		return errActionNotAuthorized
	}

	return nil
}

func (c *Controller) parseList(r *http.Request, token string) (*model.List, error) {
	var list model.List
	body, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Warn("list body cannot be read", "error", err)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	if err = json.Unmarshal(body, &list); err != nil {
		slog.Warn("list cannot be unmarshalled", "error", err)
		return nil, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	if err = c.authorizeUserID(token, list.UserID); err != nil {
		return nil, err
	}

	return &list, nil
}

func validateList(list *model.List) error {
	switch {
	case list.UserID == 0:
		return fmt.Errorf("%w: list user_id is absent", errInvalidArguments)
	case list.Name == "":
		return fmt.Errorf("%w: list name is absent", errInvalidArguments)
	}

	return nil
}
