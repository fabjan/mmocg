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
	"context"
	_ "embed" // for embedding app version
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/rs/cors"
	"github.com/uptrace/uptrace-go/uptrace"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"

	"github.com/fabjan/mmocg/server"
	"github.com/fabjan/mmocg/spam"
	"github.com/fabjan/mmocg/store"
)

type appConfig struct {
	port           int
	allowedOrigins stringSlice
	secrets        appSecrets
}

type appSecrets struct {
	uptraceDSN  string
	databaseURL string
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

func (cfg *appConfig) importSecrets() {
	// TODO secrets handling
	dsn := os.Getenv("UPTRACE_DSN")
	if dsn != "" {
		cfg.secrets.uptraceDSN = dsn
		log.Printf("\tUptrace DSN configured")
	}

	url := os.Getenv("DATABASE_URL")
	if url != "" {
		cfg.secrets.databaseURL = url
		log.Printf("\tDatabase URL configured")
	}
}

//go:embed VERSION
var appVersion string

func main() {

	appVersion = strings.TrimSpace(appVersion)

	log.Printf("Server version " + appVersion + " booting up...")

	var cfg appConfig

	log.Printf("Reading config...")

	cfg.importEnv()
	cfg.importArgs()
	cfg.importSecrets()

	log.Printf("Setting up tracing...")
	ctx := context.Background()
	uptrace.ConfigureOpentelemetry(&uptrace.Config{
		ServiceName:    "mmocg",
		ServiceVersion: appVersion,
		DSN:            cfg.secrets.uptraceDSN,
	})
	defer uptrace.Shutdown(ctx)

	// TODO We could have a buffer on this channels,
	//      but perhaps some rate limit is the first step.
	onNewTeam := make(chan string)
	onNewLeader := make(chan string)

	// TODO configurable backing implementation
	log.Printf("Setting up store...")

	var st server.Store
	st = store.NewMutMap(onNewTeam, onNewLeader)
	if cfg.secrets.databaseURL != "" {
		db, err := store.OpenPg(cfg.secrets.databaseURL)
		if err != nil {
			log.Fatalf("cannot open database connection")
		}
		st, err = store.NewPostgres(db, "teams")
		if err != nil {
			log.Fatalf("cannot initialize team store")
		}
		log.Printf("\tUsing Postgres")
	} else {
		log.Printf("\tUsing in-memory map")
	}
	defer st.Close()

	log.Printf("Setting up notification spammer...")

	// TODO graceful shutdown
	spammer := spam.NewHandler(onNewTeam, onNewLeader)
	go spammer.Go()

	log.Printf("Creating API handlers...")

	api := server.NewAPI(st)
	router := server.NewRouter(&api)
	router.Use(otelmux.Middleware("mmocg-http"))
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
