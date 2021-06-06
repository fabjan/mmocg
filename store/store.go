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
	onNewTeam   chan string
	onNewLeader chan string

	mutex sync.RWMutex
	teams map[string]server.Team
}

// NewMutMap creates a new empty MutMap.
func NewMutMap(onNewTeam, onNewLeader chan string) *MutMap {
	mm := MutMap{
		onNewTeam:   onNewTeam,
		onNewLeader: onNewLeader,
	}
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

	if mm.onNewTeam != nil {
		mm.onNewTeam <- teamID
	}

	return team, nil
}

// GetLeaderboard returns the highest scoring teams.
func (mm *MutMap) GetLeaderboard() server.Leaderboard {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	leaderboard := server.Leaderboard{}

	for _, team := range mm.teams {
		if 0 < team.Clicks {
			leaderboard = append(leaderboard, team)
		}
	}

	sort.Slice(leaderboard, func(i, j int) bool {
		// more is less
		return leaderboard[i].Clicks > leaderboard[j].Clicks
	})

	return leaderboard
}

// RecordClicks stores clicks for the given team.
func (mm *MutMap) RecordClicks(teamID string, count int64) (server.Team, error) {
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	// OnNewLeader: stash leader before updating clicks so we can know if it changed
	var prevLeader server.Team
	if mm.onNewLeader != nil {
		prevLeader = mm.lockedFindLeader()
	}

	team, ok := mm.teams[teamID]
	if !ok {
		return team, errors.New("not found")
	}

	team.Clicks += count
	mm.teams[teamID] = team

	// OnNewLeader: notify on new leader
	if mm.onNewLeader != nil {
		if prevLeader.Clicks < team.Clicks && prevLeader.ID != teamID {
			mm.onNewLeader <- teamID
		}
	}

	return team, nil
}

// locked as in you need to hold the lock when calling
// returns a "zero" team if no leader is found
func (mm *MutMap) lockedFindLeader() server.Team {
	mostPoints := int64(-1)
	leader := server.Team{}
	for _, t := range mm.teams {
		if mostPoints < t.Clicks {
			leader = t
		}
	}
	return leader
}
