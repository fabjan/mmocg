// Copyright 2021 Fabian Bergström
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"

	"github.com/fabjan/mmocg/server"
	"github.com/fabjan/mmocg/store"
)

func main() {
	log.Printf("Server started")

	// TODO configurable backing implementation
	store := store.NewMutMap()

	api := server.NewAPI(store)

	router := server.NewRouter(&api)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// TODO configurable CORS?
	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:*"},
	})

	log.Fatal(http.ListenAndServe(":"+port, c.Handler(router)))
}
