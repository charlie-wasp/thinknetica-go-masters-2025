package db

import (
	"context"

	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/models"
)

type DB interface {
	AddReview(context.Context, models.Review) error
	ListReviews(context.Context) ([]models.Review, error)
	ListReviewsByBeerID(context.Context, int) ([]models.Review, error)
	AvgBeerRating(context.Context, int) (float64, error)
}
