package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/config"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/db"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/errs"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/metrics"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/models"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/telemetry"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Server struct {
	cfg    *config.Cfg
	router *chi.Mux
	server *http.Server
	db     db.DB
}

type ErrorResponse struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
}

var err500 = errors.New("внутренняя ошибка сервера")

func New(cfg *config.Cfg, storage db.DB) *Server {
	r := chi.NewRouter()

	s := Server{
		cfg:    cfg,
		router: r,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%v", cfg.Port),
			Handler:      r,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  15 * time.Second,
		},
		db: storage,
	}

	s.endpoints()

	return &s
}

func (s *Server) endpoints() {
	// Настройка middleware
	s.router.Use(
		middleware.RequestID,                 // Добавляет X-Request-Id в заголовки
		telemetry.TracingMiddleware,          // OpenTelemetry трейсинг
		metrics.PrometheusMiddleware,         // Метрики Prometheus
		RequestLoggerMiddleware(&log.Logger), // Логирование запросов
		middleware.Recoverer,                 // Восстановление после паник
	)

	// HealthCheck - статус системы
	s.router.Get("/health", healthHandler)

	// Эндпоинт для Prometheus
	s.router.Get("/metrics", promhttp.Handler().ServeHTTP)

	// Инициализация маршрутов
	s.router.Post("/review", s.addReviewHandler)
	s.router.Get("/reviews", s.listReviewsHandler)
	s.router.Get("/avg_rating", s.avgRatingHandler)
}

func (s *Server) Start(ctx context.Context) error {
	log.Info().Msg("Инициализация телеметрии")

	shutdown, err := telemetry.SetupOTelSDK(ctx, "http://localhost:4318")
	if err != nil {
		return err
	}
	defer func() {
		err = shutdown(ctx)
		if err != nil {
			log.Err(err).Msg("cannot shutdown OTel")
		}
	}()

	log.Info().Str("addr", s.server.Addr).Msg("Запуск HTTP сервера")

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		log.Info().Msg("Остановка HTTP сервера")
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("Ошибка при остановке сервера")
		}
	}()

	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Обработчики запросов
func healthHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (s *Server) addReviewHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	defer span.End()

	log.Info().Msg("Обработка запроса addReview")
	span.AddEvent("Обработка запроса addReview")

	var req models.Review
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		span.SetStatus(codes.Error, "не удалось декодировать запрос")
		log.Err(err).Send()
		s.WriteError(w, r, errs.NewErrBadRequest("не удалось декодировать запрос"))
		return
	}

	err = s.db.AddReview(r.Context(), req)
	if err != nil {
		span.SetStatus(codes.Error, "не удалось добавить отзыв в БД")
		log.Err(err).Send()
		s.WriteError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) listReviewsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	log.Info().Msg("Обработка запроса listBeerReviews")
	span.AddEvent("Обработка запроса listBeerReviews")

	var reviews []models.Review
	var dbQueryErr error

	beerID, err := getQueryParamInt(r, "beer_id")
	if err != nil {
		s.WriteError(w, r, errs.NewErrBadRequest("невалидный beer id"))
		return
	}

	if beerID == nil {
		reviews, dbQueryErr = s.db.ListReviews(ctx)
	} else {
		reviews, dbQueryErr = s.db.ListReviewsByBeerID(ctx, *beerID)
	}

	if dbQueryErr != nil {
		span.SetStatus(codes.Error, "не удалось получить отзывы")
		log.Err(dbQueryErr).Send()
		s.WriteError(w, r, err500)
		return
	}

	s.WriteJSON(w, r, reviews)
}

func (s *Server) avgRatingHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	defer span.End()

	log.Info().Msg("Обработка запроса avgRating")
	span.AddEvent("Обработка запроса avgRating")

	beerIDString := r.URL.Query().Get("beer_id")

	if beerIDString == "" {
		s.WriteError(w, r, errs.NewErrBadRequest("отсутствует id пива"))
		return
	}

	beerID, err := strconv.Atoi(beerIDString)
	if err != nil {
		s.WriteError(w, r, errs.NewErrBadRequest("id отзыва должен быть числом"))
		return
	}

	avgRating, err := s.db.AvgBeerRating(ctx, beerID)
	if err != nil {
		span.SetStatus(codes.Error, "не удалось рассчитать средний рейтинг")
		s.WriteError(w, r, err)
		return
	}

	s.WriteJSON(w, r, struct {
		AvgRating float64 `json:"avg_rating"`
	}{avgRating})
}

func (s *Server) WriteError(w http.ResponseWriter, r *http.Request, err error) {
	resp := ErrorResponse{Message: err.Error()}
	w.Header().Set("Content-Type", "application/json")

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

func (s *Server) WriteJSON(w http.ResponseWriter, r *http.Request, response any) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		s.WriteError(w, r, err)
	}
}

// RequestLoggerMiddleware - middleware для логирования запросов
func RequestLoggerMiddleware(logger *zerolog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			logger.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Str("user_agent", r.UserAgent()).
				Int("status", ww.Status()).
				Int("bytes", ww.BytesWritten()).
				Dur("duration", time.Since(start)).
				Str("request_id", middleware.GetReqID(r.Context())).
				Msg("Обработан HTTP запрос")
		})
	}
}

func getQueryParamInt(r *http.Request, paramName string) (*int, error) {
	paramValString := r.URL.Query().Get(paramName)

	if paramValString == "" {
		return nil, nil
	}

	paramVal, err := strconv.Atoi(paramValString)
	if err != nil {
		return nil, err
	}

	return &paramVal, nil
}
