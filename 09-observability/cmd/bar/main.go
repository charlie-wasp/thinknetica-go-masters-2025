package main

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/charlie-wasp/go-masters-2025/observability/internal/logger"
	"github.com/charlie-wasp/go-masters-2025/observability/internal/random"
	"github.com/charlie-wasp/go-masters-2025/observability/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	logger.Init(zerolog.DebugLevel)
	router := chi.NewRouter()

	router.Use(telemetry.TracingMiddleware)

	router.Get("/bar", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		span := trace.SpanFromContext(ctx)
		defer span.End()

		sleepDuration := random.Duration(800)
		log.Info().Msgf("imitating some work for %s", sleepDuration)
		time.Sleep(sleepDuration)

		if rand.Int()%2 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			span.SetStatus(codes.Error, "test error")
			return
		}

		w.Write([]byte("bar"))
	})

	telemetry.SetupOTelSDK(context.TODO(), "http://localhost:4318", "bar-service")

	listenAddr := ":8081"
	server := http.Server{Addr: listenAddr, Handler: router}

	log.Info().Msgf("Bar service is listening on %s", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Err(err).Send()
	}
}
