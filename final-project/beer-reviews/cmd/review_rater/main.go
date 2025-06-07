package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/config"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/db/postgres"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/ollama"
	"github.com/rs/zerolog/log"
)

const (
	prompt = "rate this review from 1 to 5 and respond with single number only: %s"
)

var noDataWaitDuration = 10 * time.Second

func main() {
	cfg, err := config.Load(".")
	if err != nil {
		log.Fatal().Err(err).Msg("Ошибка при загрузке конфигурации")
	}

	db, err := postgres.New(cfg.DBConnStr)
	if err != nil {
		log.Fatal().Msgf("ошибка инициализации БД: %s", err)
	}

	ollamaClient, err := ollama.NewClient(cfg.OllamaURL)
	if err != nil {
		log.Fatal().Msgf("ошибка инициализации клиента Ollama: %s", err)
	}

	for {
		reviews, err := db.ListReviewsWithoutRating(context.TODO())
		if err != nil {
			log.Err(err).Msg("ошибка при чтении отзывов без рейтинга из БД")
			continue
		}

		log.Info().Msgf("найдено %d отзывов без рейтинга", len(reviews))

		if len(reviews) == 0 {
			time.Sleep(noDataWaitDuration)
			continue
		}

		for _, review := range reviews {
			log.Info().Int("review_id", review.ID).Msg("запрашиваем Ollama...")
			ollamaResponse, err := ollamaClient.Query(cfg.OllamaModel, fmt.Sprintf(prompt, review.Content))
			if err != nil {
				log.Err(err).Int("review_id", review.ID).Msg("ошибка при запросе Ollama")
				continue
			}
			rating, err := strconv.Atoi(ollamaResponse)
			if err != nil {
				log.Err(err).Int("review_id", review.ID).Msgf("Ollama вернула неожиданный результат '%s'", ollamaResponse)
				continue
			}
			review.Rating = &rating
			err = db.UpdateReview(context.TODO(), review)
			if err != nil {
				log.Err(err).Int("review_id", review.ID).Msg("ошибка при попытке обновить отзыв")
				continue
			}
			log.Info().Msgf("отзыв с id=%d успешно обновлен", review.ID)
		}
	}
}
