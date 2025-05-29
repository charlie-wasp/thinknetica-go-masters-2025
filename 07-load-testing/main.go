package main

import (
	"log"
	"net/http"

	"github.com/valyala/fasthttp"
)

const responseText = "Hello to you too!"

func main() {
	go func() {
		fasthttpHandler := func(ctx *fasthttp.RequestCtx) {
			switch string(ctx.Path()) {
			case "/hello":
				ctx.WriteString(responseText)
			default:
				ctx.Error("Not Found", fasthttp.StatusNotFound)
			}
		}

		log.Println("Fasthttp server listening on :8081")
		if err := fasthttp.ListenAndServe(":8081", fasthttpHandler); err != nil {
			log.Printf("Fasthttp server error: %v\n", err)
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(responseText))
	})

	log.Println("Standard HTTP server listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Printf("Standard server error: %v", err)
	}
}
