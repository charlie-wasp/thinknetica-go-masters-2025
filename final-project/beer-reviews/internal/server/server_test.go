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

func TestFibonacciHandler(t *testing.T) {
	dbMock := memdb.New()
	dbMock.AddReview(
		context.Background(),
		models.Review{
			ID:      1,
			Content: "Test content",
			UserID:  1,
			BeerID:  1,
			Rating:  &[]int{3}[0],
		},
	)

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
			expectedBody:   `[{"id":1,"content":"Test content","rating":3,"user_id":1,"beer_id":1}]`,
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
