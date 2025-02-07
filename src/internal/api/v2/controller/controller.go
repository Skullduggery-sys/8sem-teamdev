package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"git.iu7.bmstu.ru/vai20u117/testing/src/internal/service"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Controller struct {
	poster        *PosterHandler
	list          *ListHandler
	listPoster    *ListPosterHandler
	posterHistory *PosterRecordHandler
	auth          *AuthHandler
}

func NewController(
	poster *PosterHandler,
	list *ListHandler,
	listPoster *ListPosterHandler,
	posterHistory *PosterRecordHandler,
	auth *AuthHandler,
) *Controller {
	return &Controller{
		poster:        poster,
		list:          list,
		listPoster:    listPoster,
		posterHistory: posterHistory,
		auth:          auth,
	}
}

func (c *Controller) CreateRouter(router *mux.Router) *mux.Router {
	// /swagger/index.html
	router.PathPrefix("/swagger").Handler(httpSwagger.WrapHandler)

	router.HandleFunc("/get-user-token", c.handleGetUserTokenRequests).
		Methods(http.MethodPost)
	router.HandleFunc("/sign-up", c.handleSignUpRequests).
		Methods(http.MethodPost)
	router.HandleFunc("/sign-in", c.handleSignInRequests).
		Methods(http.MethodPost)
	router.HandleFunc("/sign-out", c.handleSignOutRequests).
		Methods(http.MethodPost)

	router.HandleFunc("/posters/{id}", c.handlePosterPathRequests).
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	router.HandleFunc("/posters", c.handlePosterBodyRequests).
		Methods(http.MethodPost)

	router.HandleFunc("/lists/{id}", c.handleListPathRequests).
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete)
	router.HandleFunc("/lists-root", c.handleListUserRequests).
		Methods(http.MethodGet)
	router.HandleFunc("/sublists/{id}", c.handleSublistsRequests).
		Methods(http.MethodGet)
	router.HandleFunc("/lists", c.handleListDefaultRequests).
		Methods(http.MethodPost)

	router.HandleFunc("/lists/{list_id}/posters/{poster_id}", c.handleListPosterRequests).
		Methods(http.MethodPost, http.MethodPut, http.MethodDelete)
	router.HandleFunc("/lists/{list_id}/posters", c.handleGetListPostersRequests).
		Methods(http.MethodGet)

	router.HandleFunc("/poster-records/{poster_id}", c.handlePosterRecordPathRequests).
		Methods(http.MethodPost, http.MethodDelete)
	router.HandleFunc("/poster-records", c.handlePosterRecordDefaultRequests).
		Methods(http.MethodGet)
	return router
}

func (c *Controller) getUserIDByToken(token string) (int, error) {
	userID, err := c.auth.service.GetUserID(token)
	if errors.Is(err, service.ErrNotFound) {
		slog.Warn("user_id by token not found", "token", token)
		return 0, errUserNotFound
	} else if err != nil {
		slog.Error("failed to get user_id by token", "error", err)
		return 0, err
	}

	return userID, nil
}

// TODO: rename to writeErrorCode.
func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errInvalidArguments):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, errNotFound):
		w.WriteHeader(http.StatusNotFound)
	case errors.Is(err, errInternal):
		w.WriteHeader(http.StatusInternalServerError)
	case errors.Is(err, errUserNotFound):
		w.WriteHeader(http.StatusBadRequest)
	case errors.Is(err, errActionNotAuthorized):
		w.WriteHeader(http.StatusUnauthorized)
	default:
		w.WriteHeader(http.StatusTeapot)
		slog.Error("type of error is unknown to controller, returning teapot status", "error", err)
	}

	errJSON, errMarshal := json.Marshal(map[string]string{"error": err.Error()})
	if errMarshal != nil {
		slog.Error("cannot marshal error", "error", err, "marshal_error", errMarshal)
	}

	if _, err := w.Write(errJSON); err != nil {
		slog.Error("unexpected error occurred while write marshaled error", "error", err)
	}
}

func parseInt(r *http.Request, argName string) (int, error) {
	idReq := mux.Vars(r)[argName]
	id, err := strconv.ParseInt(idReq, 10, 64)
	if err != nil {
		slog.Warn("cannot convert to int", "arg_name", argName, "arg_value", idReq)
		return 0, fmt.Errorf("%w: %w", errInvalidArguments, err)
	}

	return int(id), nil
}
