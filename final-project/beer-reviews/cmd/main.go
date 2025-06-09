package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/config"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/db/postgres"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/server"

	"github.com/rs/zerolog/log"
)

func main() {
	// Создаем контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Инициализируем конфигурацию
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Ошибка при загрузке конфигурации")
	}

	// Инициализируем подключение к БД
	db, err := postgres.New(cfg.DBConnStr)
	if err != nil {
		log.Fatal().Err(err).Msg("ошибка инициализации БД")
	}

	// Инициализируем сервер
	srv := server.New(cfg, db)

	// Запускаем сервер в отдельной горутине
	go func() {
		if err := srv.Start(ctx); err != nil {
			log.Error().Err(err).Msg("Ошибка при запуске сервера")
			cancel()
		}
	}()

	// Ожидаем сигналы ОС для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Получен сигнал на завершение работы")
	cancel()

	// Ожидаем завершения работы сервера
	<-ctx.Done()
	log.Info().Msg("Сервер успешно остановлен")
}
