package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/config"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/db/memdb"
	"github.com/charlie-wasp/go-masters-2025/beer-reviews/internal/models"
)

var testReviews = []models.Review{
	{
		ID:      1,
		Content: "Test content",
		UserID:  1,
		BeerID:  1,
		Rating:  &[]int{3}[0],
	},
	{
		ID:      2,
		Content: "Beer 2 review",
		UserID:  1,
		BeerID:  2,
		Rating:  &[]int{4}[0],
	},
}

func TestReviewsHandler(t *testing.T) {
	dbMock := memdb.New()
	for _, r := range testReviews {
		dbMock.AddReview(context.Background(), r)
	}

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Get all reviews",
			url:            "/reviews",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"content":"Test content","rating":3,"user_id":1,"beer_id":1},{"id":2,"content":"Beer 2 review","rating":4,"user_id":1,"beer_id":2}]`,
		},
		{
			name:           "Get reviews of particular beer",
			url:            "/reviews?beer_id=2",
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":2,"content":"Beer 2 review","rating":4,"user_id":1,"beer_id":2}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			api := New(&config.Cfg{}, dbMock)
			rr := httptest.NewRecorder()

			api.router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if strings.TrimSpace(rr.Body.String()) != tt.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
