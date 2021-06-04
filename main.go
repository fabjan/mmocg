package main

import (
	"log"
	"net/http"
	"os"

	"github.com/fabjan/mmocg/server"
)

func main() {
	log.Printf("Server started")

	router := server.NewRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(http.ListenAndServe(":"+port, router))
}
