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

// Package store persists game data.
package store

import (
	"errors"
	"sort"
	"sync"

	"github.com/fabjan/mmocg/server"
)

// MutMap is an in memory team score store.
type MutMap struct {
	mutex sync.RWMutex
	teams map[string]server.Team
}

// NewMutMap creates a new empty MutMap.
func NewMutMap() *MutMap {
	mm := MutMap{}
	mm.teams = make(map[string]server.Team)
	return &mm
}

// FindByID returns a single team if found by ID.
func (mm *MutMap) FindByID(teamID string) (server.Team, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	team, ok := mm.teams[teamID]
	if !ok {
		return team, errors.New("not found")
	}

	return team, nil
}

// CreateTeam creates a new team, an error means the ID is taken.
func (mm *MutMap) CreateTeam(teamID string) (server.Team, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	team, ok := mm.teams[teamID]
	if ok {
		return team, errors.New("exists")
	}

	// If the team was not found, we insert
	// a "zeroed" team with just the ID.
	team.ID = teamID
	mm.teams[teamID] = team

	return team, nil
}

// GetLeaderboard returns the highest scoring teams.
func (mm *MutMap) GetLeaderboard() server.Leaderboard {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	leaderboard := server.Leaderboard{}

	for _, team := range mm.teams {
		leaderboard = append(leaderboard, team)
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		return leaderboard[i].Clicks < leaderboard[j].Clicks
	})

	return leaderboard
}

// RecordClicks stores clicks for the given team.
func (mm *MutMap) RecordClicks(teamID string, count int64) (server.Team, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	team, ok := mm.teams[teamID]
	if !ok {
		return team, errors.New("not found")
	}

	team.Clicks += count
	mm.teams[teamID] = team

	return team, nil
}