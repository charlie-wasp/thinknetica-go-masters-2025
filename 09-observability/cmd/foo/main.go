package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/charlie-wasp/go-masters-2025/observability/internal/logger"
	"github.com/charlie-wasp/go-masters-2025/observability/internal/random"
	"github.com/charlie-wasp/go-masters-2025/observability/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

func main() {
	logger.Init(zerolog.DebugLevel)

	router := chi.NewRouter()
	router.Use(telemetry.TracingMiddleware)

	router.Get("/foo", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		span := trace.SpanFromContext(ctx)
		defer span.End()

		sleepDuration := random.Duration(800)
		log.Info().Msgf("imitating some work for %s", sleepDuration)
		time.Sleep(sleepDuration)

		barServerResponse, err := requestBarService(ctx)
		if err != nil {
			log.Err(err).Send()
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("failed to request bar service"))
			return
		}

		w.Write(append([]byte("foo"), barServerResponse...))
	})

	telemetry.SetupOTelSDK(context.TODO(), "http://localhost:4318", "foo-service")

	listenAddr := ":8080"
	server := http.Server{Addr: listenAddr, Handler: router}

	log.Info().Msgf("Foo service is listening on %s", listenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Err(err).Send()
	}
}

func requestBarService(ctx context.Context) ([]byte, error) {
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8081/bar", nil)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to build request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to request bar service: %v", err)
	}
	defer resp.Body.Close()

	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to read bar service response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("bar service responded with status %d and body '%s'", resp.StatusCode, response)
	}

	return response, nil
}
