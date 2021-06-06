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
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rs/cors"

	"github.com/fabjan/mmocg/server"
	"github.com/fabjan/mmocg/spam"
	"github.com/fabjan/mmocg/store"
)

type appConfig struct {
	port           int
	allowedOrigins stringSlice
}

func (cfg *appConfig) importEnv() {
	envPort := os.Getenv("PORT")
	port, err := strconv.Atoi(envPort)
	if err != nil {
		port = 5000
	}
	cfg.port = port
	log.Printf("\tAPI port: %d", cfg.port)
}

func (cfg *appConfig) importArgs() {

	var flagAllowOrigins stringSlice
	flag.Var(&flagAllowOrigins, "allow-origin", "Patterns to allow as origin in CORS.")

	flag.Parse()

	cfg.allowedOrigins = flagAllowOrigins

	log.Printf("\tAllowed origins: %s", cfg.allowedOrigins.String())
}

func main() {

	var cfg appConfig

	log.Printf("Reading config...")

	cfg.importEnv()
	cfg.importArgs()

	// TODO We could have a buffer on this channels,
	//      but perhaps some rate limit is the first step.
	onNewTeam := make(chan string)
	onNewLeader := make(chan string)

	// TODO configurable backing implementation
	log.Printf("Setting up store...")

	store := store.NewMutMap(onNewTeam, onNewLeader)

	log.Printf("Setting up notification spammer...")

	// TODO graceful shutdown
	spammer := spam.NewHandler(onNewTeam, onNewLeader)
	go spammer.Go()

	log.Printf("Creating API handlers...")

	api := server.NewAPI(store)
	router := server.NewRouter(&api)
	c := cors.New(cors.Options{
		AllowedOrigins: cfg.allowedOrigins,
	})

	log.Printf("Server is listening...")

	addr := fmt.Sprintf(":%d", cfg.port)
	log.Fatal(http.ListenAndServe(addr, c.Handler(router)))
}

type stringSlice []string

func (s *stringSlice) String() string {
	var sb strings.Builder
	sb.WriteString("[")
	sb.WriteString(strings.Join(*s, ", "))
	sb.WriteString("]")
	return fmt.Sprintf("%v", sb.String())
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}
