package model

import (
	"time"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
	"github.com/samber/lo"
)

type PosterRecordResponse struct {
	ID        int       `json:"id"`
	PosterID  int       `json:"posterId"`
	UserID    int       `json:"userId"`
	CreatedAt time.Time `json:"createdat"`
}

func ToPosterRecordResponse(records []*model.PosterRecord) []*PosterRecordResponse {
	return lo.Map(records, func(item *model.PosterRecord, _ int) *PosterRecordResponse {
		return &PosterRecordResponse{
			ID:        item.ID,
			PosterID:  item.PosterID,
			UserID:    item.UserID,
			CreatedAt: item.CreatedAt,
		}
	})
}
