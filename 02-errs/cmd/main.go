package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charlie-wasp/go-masters-2025/errors-server/internal/api"
	"github.com/charlie-wasp/go-masters-2025/errors-server/internal/logger"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	logger.Init(zerolog.DebugLevel)
	port := 8080

	s := api.New(port)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill, syscall.SIGTERM)
	defer cancel()

	go func() {
		log.Info().Msgf("server is listening to port %d", port)
		err := s.Serve(ctx)
		if err != nil {
			log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()
	s.Shutdown()
}
