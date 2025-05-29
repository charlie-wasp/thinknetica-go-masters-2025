package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFibonacciHandler(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Valid input 0",
			url:            "/fibo?N=0",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"input":0,"result":0}`,
		},
		{
			name:           "Valid input 1",
			url:            "/fibo?N=1",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"input":1,"result":1}`,
		},
		{
			name:           "Valid input 11",
			url:            "/fibo?N=11",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"input":11,"result":89}`,
		},
		{
			name:           "Missing parameter",
			url:            "/fibo",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"statusCode":400,"message":"missing parameter 'N'"}`,
		},
		{
			name:           "Invalid parameter (non-number)",
			url:            "/fibo?N=abc",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"statusCode":400,"message":"parameter 'N' must be an integer"}`,
		},
		{
			name:           "Negative number",
			url:            "/fibo?N=-5",
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"statusCode":400,"message":"parameter 'N' must be a non-negative integer"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.url, nil)
			if err != nil {
				t.Fatal(err)
			}

			api := New(WithEnv(TestEnv))
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(api.fibonacciHandler)

			handler.ServeHTTP(rr, req)

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
