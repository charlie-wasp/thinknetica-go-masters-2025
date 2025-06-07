package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/utc", func(w http.ResponseWriter, r *http.Request) {
		currentTime := time.Now().UTC().Format("2006-01-02 15:04:05 UTC")
		fmt.Fprint(w, currentTime)
	})

	port := ":8000"
	fmt.Printf("Server running on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
