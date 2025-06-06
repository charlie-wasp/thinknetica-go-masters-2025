package memdb

import (
	"context"

	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/models"
)

type MemDB struct {
	data []models.Review
}

func New() *MemDB {
	return &MemDB{}
}

func (m *MemDB) AddReview(_ context.Context, review models.Review) error {
	m.data = append(m.data, review)
	return nil
}

func (m *MemDB) ListReviews(_ context.Context) ([]models.Review, error) {
	return m.data, nil
}

func (m *MemDB) ListReviewsByBeerID(_ context.Context, beerID int) ([]models.Review, error) {
	res := make([]models.Review, 0)

	for i, r := range m.data {
		if r.BeerID != beerID {
			continue
		}
		res = append(res, m.data[i])
	}
	return res, nil
}

func (m *MemDB) ListReviewsByUserID(_ context.Context, userID int) ([]models.Review, error) {
	res := make([]models.Review, 0)

	for i, r := range m.data {
		if r.UserID != userID {
			continue
		}
		res = append(res, m.data[i])
	}
	return res, nil
}

func (m *MemDB) AvgBeerRating(_ context.Context, beerID int) (float64, error) {
	var res float64
	var dataLen int

	for _, r := range m.data {
		if r.Rating == nil || r.BeerID != beerID {
			continue
		}
		dataLen++
		res += float64(*r.Rating)
	}
	// наивная имплементация, скорее всего будут проблемы
	return res / float64(dataLen), nil
}
