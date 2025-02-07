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

// @Summary	List all user records
// @Description	lists all user records
// @Tags poster-records/v2
// @Param X-User-Token header string true "JWT-format token"
// @Accept json
// @Success	200 {array} reqModelPkg.PosterRecordResponse
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/poster-records [get]
func (h *PosterRecordHandler) GetUserRecords(ctx context.Context, userID int) ([]byte, error) {
	records, err := h.service.GetUserRecords(ctx, userID)
	if errors.Is(err, servicePkg.ErrNotFound) {
		slog.Warn("user recoreds not found", "user_id", userID)
		return nil, fmt.Errorf("%w: %w", errNotFound, err)
	} else if err != nil {
		slog.Error("unexpected error occurred while getting all history records", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	recordsResponse := reqModelPkg.ToPosterRecordResponse(records)
	recordsJSON, err := json.Marshal(recordsResponse)
	if err != nil {
		slog.Error("unexpected error occurred while marshaling history records", "error", err)
		return nil, fmt.Errorf("%w: %w", errInternal, err)
	}

	return recordsJSON, nil
}

// @Summary	Create poster record
// @Description	create poster record
// @Tags poster-records/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param poster_id path integer true "PosterId"
// @Accept json
// @Success	201 {object} reqModelPkg.IDResponse "id"
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/poster-records/{poster_id} [post]
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
// @Tags poster-records/v2
// @Param X-User-Token header string true "JWT-format token"
// @Param poster_id path integer true "PosterId"
// @Accept json
// @Success	200
// @Failure	400	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	401	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	404	{object} reqModelPkg.ErrorResponse "Error"
// @Failure	500	{object} reqModelPkg.ErrorResponse "Error"
// @Router /api/v2/poster-records/{poster_id} [delete]
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
func (c *Controller) handlePosterRecordPathRequests(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(tokenHeader)
	if authErr := c.auth.service.Authorize(token); authErr != nil {
		writeError(w, fmt.Errorf("%w: %w", errActionNotAuthorized, authErr))
		return
	}

	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		userID, err := c.getUserIDByToken(token)
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
		posterID, err := parseInt(r, "poster_id")
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

func (c *Controller) handlePosterRecordDefaultRequests(w http.ResponseWriter, r *http.Request) {
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

		records, err := c.posterHistory.GetUserRecords(ctx, userID)
		if err != nil {
			writeError(w, err)
			break
		} else if _, err = w.Write(records); err != nil {
			writeError(w, fmt.Errorf("%w: writing records body: %w", errInternal, err))
			break
		}

		w.WriteHeader(http.StatusOK)
	default:
		slog.Error("http method is not allowed", "method", r.Method)
		w.WriteHeader(http.StatusForbidden)
	}
}
