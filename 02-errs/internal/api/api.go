package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/charlie-wasp/go-masters-2025/errors-server/internal/errs"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

type API struct {
	port   int
	router *chi.Mux
}

type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func New(port int) *API {
	api := API{
		router: chi.NewRouter(),
		port:   port,
	}
	api.registerEndpoints()
	return &api
}

func (api *API) Serve(ctx context.Context) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%v", api.port),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		Handler:      api.router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	return srv.ListenAndServe()
}

func (api *API) WriteJSON(w http.ResponseWriter, r *http.Request, response any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		api.WriteError(w, r, err)
	}
}

// WriteError sends error message and status code.
func (api *API) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	resp := ErrorResponse{Message: err.Error()}
	w.Header().Set("Content-Type", "application/json")

	log.Err(err).Send()
	switch err.(type) {
	case errs.ErrNoData:
		resp.StatusCode = http.StatusNotFound
	case errs.ErrBadRequest:
		resp.StatusCode = http.StatusBadRequest
	default:
		resp.StatusCode = http.StatusInternalServerError
	}

	w.WriteHeader(resp.StatusCode)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		log.Err(err).Send()
	}
}

func (api *API) Shutdown() {
	log.Info().Msg("graceful server shutdown")
}

func (api *API) registerEndpoints() {
	api.router.Get("/users/{id}", api.usersHandler)
}

func (api *API) usersHandler(w http.ResponseWriter, r *http.Request) {
	s := chi.URLParam(r, "id")
	id, err := strconv.Atoi(s)
	if err != nil {
		e := errs.NewErrBadRequest(
			fmt.Sprintf("invalid user id '%s' is given", s),
		)
		api.WriteError(w, r, e)
		return
	}

	if id != 1 {
		e := errs.NewErrNoData(
			fmt.Sprintf("user with id %d is not found", id),
		)
		api.WriteError(w, r, e)
		return
	}

	api.WriteJSON(w, r, struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}{1, "John Doe"})
}
