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

package server

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// API uses a store to respond to API requests
type API struct {
	store Store
}

// NewAPI creates an API handler using the given store
func NewAPI(store Store) API {
	return API{store}
}

// Store stores scores and teams
type Store interface {
	// error must mean the team ID is taken
	CreateTeam(teamID string) (Team, error)
	FindByID(teamID string) (Team, error)
	GetLeaderboard() Leaderboard
	RecordClicks(teamID string, count int64) (Team, error)
	Close()
}

// UpdateTeam creates or updates a team using the form data from the request
func (api *API) UpdateTeam(w http.ResponseWriter, r *http.Request) {

	teamID, ok := mux.Vars(r)["teamId"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// TODO check if team name is really emoji, return 400 if not
	// https://stackoverflow.com/questions/30757193/find-out-if-character-in-string-is-emoji/

	setContentTypeJSON(w)

	team, err := api.store.CreateTeam(teamID)
	if err == nil {
		// this must mean the team was created
		w.WriteHeader(http.StatusCreated)
	}

	json.NewEncoder(w).Encode(team)
}

// GetTeamByID returns a single team if found by ID
func (api *API) GetTeamByID(w http.ResponseWriter, r *http.Request) {

	teamID, ok := mux.Vars(r)["teamId"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team, err := api.store.FindByID(teamID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	setContentTypeJSON(w)
	json.NewEncoder(w).Encode(team)
}

// GetLeaderboard returns the highest scoring teams
func (api *API) GetLeaderboard(w http.ResponseWriter, r *http.Request) {

	lb := api.store.GetLeaderboard()

	setContentTypeJSON(w)
	json.NewEncoder(w).Encode(lb)
}

var minCount = 1
var maxCount = 10

// Click reports clicks for the given team
func (api *API) Click(w http.ResponseWriter, r *http.Request) {

	teamID, ok := mux.Vars(r)["teamId"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	countParam := r.URL.Query().Get("count")
	if countParam == "" {
		countParam = "1"
	}

	count, err := strconv.Atoi(countParam)
	if err != nil || count < minCount || maxCount < count {
		// TODO implement pay to win
		w.WriteHeader(http.StatusPaymentRequired)
		return
	}

	// TODO rate limit -> 429

	team, err := api.store.RecordClicks(teamID, int64(count))
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	setContentTypeJSON(w)
	json.NewEncoder(w).Encode(team)
}

func setContentTypeJSON(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
}
