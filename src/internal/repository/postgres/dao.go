package postgres

import (
	"strings"
	"time"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/model"
)

type posterDAO struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	Year      int       `db:"year"`
	Genres    string    `db:"genres"`
	Chrono    int       `db:"chrono"`
	UserID    int       `db:"user_id"`
	CreatedAt time.Time `db:"created_at"`
}

type listDAO struct {
	ID       int    `db:"id"`
	ParentID int    `db:"parent_id"`
	Name     string `db:"name"`
	UserID   int    `db:"user_id"`
}

type userDAO struct {
	ID       int    `db:"id"`
	Name     string `db:"name"`
	Login    string `db:"login"`
	Role     string `db:"role"`
	Password string `db:"password"`
}

func mapListDAO(list *listDAO) *model.List {
	return &model.List{
		ID:       list.ID,
		ParentID: list.ParentID,
		Name:     list.Name,
		UserID:   list.UserID,
	}
}

func reverseMapListDAO(list *model.List) *listDAO {
	return &listDAO{
		ID:       list.ID,
		ParentID: list.ParentID,
		Name:     list.Name,
		UserID:   list.UserID,
	}
}

func mapPosterDAO(poster *posterDAO) *model.Poster {
	genres := strings.Split(poster.Genres, ",")
	return &model.Poster{
		ID:        poster.ID,
		Name:      poster.Name,
		Year:      poster.Year,
		Genres:    genres,
		Chrono:    poster.Chrono,
		UserID:    poster.UserID,
		CreatedAt: poster.CreatedAt,
	}
}

func reverseMapPosterDAO(poster *model.Poster) *posterDAO {
	genres := strings.Join(poster.Genres, ",")
	return &posterDAO{
		ID:        poster.ID,
		Name:      poster.Name,
		Year:      poster.Year,
		Genres:    genres,
		Chrono:    poster.Chrono,
		UserID:    poster.UserID,
		CreatedAt: poster.CreatedAt,
	}
}

func mapUserDAO(user *userDAO) *model.User {
	return &model.User{
		ID:       user.ID,
		Name:     user.Name,
		Login:    user.Login,
		Role:     user.Role,
		Password: user.Password,
	}
}

func reverseMapUserDAO(user *model.User) *userDAO {
	return &userDAO{
		ID:       user.ID,
		Name:     user.Name,
		Login:    user.Login,
		Role:     user.Role,
		Password: user.Password,
	}
}
