package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

const llmModel = "llama3.2"

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
}

func main() {
	targetURL := os.Getenv("TARGET_SERVER")
	if targetURL == "" {
		targetURL = "http://localhost:11434"
		log.Printf("TARGET_SERVER is not set, using default %s", targetURL)
	}

	target, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("Failed to parse target URL: %v", err)
	}
	target.Path = "/api/generate"

	client := &http.Client{Timeout: 30 * time.Second}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Received request: %s %s", r.Method, r.URL.Path)

		prompt := r.URL.Query().Get("q")

		if prompt == "" {
			http.Error(w, "Prompt is empty", http.StatusBadRequest)
			return
		}

		requestObject := ollamaRequest{
			Model:  llmModel,
			Prompt: prompt,
			Stream: false,
		}
		requestBody, err := json.Marshal(requestObject)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error marshalling request: %v", err)
			return
		}

		req, err := http.NewRequest(http.MethodPost, target.String(), bytes.NewBuffer(requestBody))
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			log.Printf("Error creating request: %v", err)
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error requesting llm server", http.StatusBadGateway)
			log.Printf("Error requesting llm server: %v", err)
			return
		}

		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		var resultBody ollamaResponse
		err = decoder.Decode(&resultBody)
		if err != nil {
			http.Error(w, "Error parsing llm response", http.StatusInternalServerError)
			log.Printf("Error parsing llm response: %v", err)
			return
		}

		fmt.Fprint(w, resultBody.Response)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on :%s, forwarding to %s", port, targetURL)

	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatal()
	}
}
