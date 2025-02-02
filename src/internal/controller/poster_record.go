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

type PosterRecordService interface {
	GetUserRecords(ctx context.Context, userID int) ([]*model.PosterRecord, error)
	CreateRecord(ctx context.Context, posterID, userID int) (int, error)
	DeleteRecord(ctx context.Context, posterID int) error
}

type PosterRecordHandler struct {
	service PosterRecordService
}

func NewPosterRecordHandler(service PosterRecordService) *PosterRecordHandler {
	return &PosterRecordHandler{service: service}
}

// @Summary	Get user records
// @Description	get user records
// @Tags poster-history
// @Param token query string true "User auth token"
// @Param user_id query integer true "UserId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /poster-history [get]
func (h *PosterRecordHandler) GetUserRecords(ctx context.Context, userID int) ([]byte, error) {
	records, err := h.service.GetUserRecords(ctx, userID)
	if err != nil {
		slog.Error("unexpected error occurred while getting all history records", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	recordsJSON, err := json.Marshal(records)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling history records", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return recordsJSON, nil
}

// @Summary	Create poster record
// @Description	create poster record
// @Tags poster-history
// @Param token query string true "User auth token"
// @Param poster_id query integer true "PosterId"
// @Param user_id query integer true "UserId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /poster-history [post]
func (h *PosterRecordHandler) CreateRecord(ctx context.Context, posterID, userID int) ([]byte, error) {
	id, err := h.service.CreateRecord(ctx, posterID, userID)
	if err != nil {
		slog.Error("unexpected error occurred while creating record", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	idJSON, err := json.Marshal(map[string]int{"id": id})
	if err != nil {
		slog.Error("unexpected error occurred while marshaling id", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return idJSON, nil
}

// @Summary	Delete user record
// @Description	delete user record
// @Tags poster-history
// @Param token query string true "User auth token"
// @Param poster_id query integer true "PosterId"
// @Accept json
// @Success	200 {object} map[string]interface{}
// @Failure	400	{object} map[string]interface{}
// @Failure	401	{object} map[string]interface{}
// @Failure	404	{object} map[string]interface{}
// @Failure	500	{object} map[string]interface{}
// @Router /poster-history [delete]
func (h *PosterRecordHandler) DeleteRecord(ctx context.Context, posterID int) error {
	if err := h.service.DeleteRecord(ctx, posterID); errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("record not found", "record_id", posterID)
		return errNotFound
	} else if err != nil {
		slog.Error("unexpected error occurred while deleting record", "error", err)
		return fmt.Errorf("%w: %w", errInternal, err)
	}

	return nil
}

//nolint:cyclop // http handler methods router
func (c *Controller) handlePosterRecordRequests(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	ctx := r.Context()
	switch r.Method {
	case http.MethodGet:
		userID, err := c.getUserIDByToken(token)
		if err != nil {
			writeError(w, err)
			break
		}

		records, err := c.posterHistory.GetUserRecords(ctx, userID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(records); err != nil {
			writeError(w, fmt.Errorf("%w: writing records body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	case http.MethodPost:
		userID, err := c.getUserIDByToken(token)
		if err != nil {
			writeError(w, err)
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

		id, err := c.posterHistory.CreateRecord(ctx, posterID, userID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(id); err != nil {
			writeError(w, fmt.Errorf("%w: writing record id body: %w", errInternal, err))
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

		if err := c.posterHistory.DeleteRecord(ctx, posterID); err != nil {
			writeError(w, err)
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}
