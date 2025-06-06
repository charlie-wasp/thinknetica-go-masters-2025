package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/charlie-wasp/go-masters-2025/fibonacci-server/internal/errs"
	"github.com/charlie-wasp/go-masters-2025/fibonacci-server/internal/fibonacci"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

const TestEnv = "test"

type API struct {
	port   int
	env    string
	router *chi.Mux
}

type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

func New(options ...func(*API)) *API {
	api := &API{router: chi.NewRouter()}
	for _, option := range options {
		option(api)
	}
	api.registerEndpoints()
	return api
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

func (api *API) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	resp := ErrorResponse{Message: err.Error()}
	w.Header().Set("Content-Type", "application/json")

	api.logError(err)
	switch err.(type) {
	case errs.ErrBadRequest:
		resp.StatusCode = http.StatusBadRequest
	default:
		resp.StatusCode = http.StatusInternalServerError
	}

	w.WriteHeader(resp.StatusCode)
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		api.logError(err)
	}
}

func (api *API) Shutdown() {
	log.Info().Msg("graceful server shutdown")
}

func (api *API) registerEndpoints() {
	api.router.Get("/fibo", api.fibonacciHandler)
}

func (api *API) logError(err error) {
	if api.env == TestEnv {
		return
	}
	log.Err(err).Send()
}

func (api *API) fibonacciHandler(w http.ResponseWriter, r *http.Request) {
	nStr := r.URL.Query().Get("N")
	if nStr == "" {
		e := errs.NewErrBadRequest("missing parameter 'N'")
		api.WriteError(w, r, e)
		return
	}

	n, err := strconv.Atoi(nStr)
	if err != nil {
		e := errs.NewErrBadRequest("parameter 'N' must be an integer")
		api.WriteError(w, r, e)
		return
	}

	if n < 0 {
		e := errs.NewErrBadRequest("parameter 'N' must be a non-negative integer")
		api.WriteError(w, r, e)
		return
	}

	result := fibonacci.Calculate(uint(n))

	api.WriteJSON(w, r, struct {
		Input  int    `json:"input"`
		Result uint64 `json:"result"`
	}{n, result})
}

func WithPort(port int) func(*API) {
	return func(api *API) {
		api.port = port
	}
}

func WithEnv(env string) func(*API) {
	return func(api *API) {
		api.env = env
	}
}
