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

package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib" // for sql.Open("pgx", ...)

	"github.com/fabjan/mmocg/server"
)

// Postgres is a postrges backed team score store.
type Postgres struct {
	db          *sql.DB
	tableName   string
	onNewTeam   chan string
	onNewLeader chan string
}

// OpenPg opens a connection to the Postgres database with the given URL.
func OpenPg(rawURL string) (*sql.DB, error) {
	return sql.Open("pgx", rawURL)
}

// NewPostgres creates a Postgres backed by the given table and DB.
// The table is created if it does not exist.
func NewPostgres(db *sql.DB, name string, onNewTeam, onNewLeader chan string) (*Postgres, error) {
	s := Postgres{
		tableName: name,
		db:        db,
	}

	_, err := db.Exec(s.createTableSQL())
	if err != nil {
		return nil, err
	}

	s.onNewTeam = onNewTeam
	s.onNewLeader = onNewLeader

	return &s, nil
}

// Close closes the store (its database connection)
func (s *Postgres) Close() {
	s.db.Close()
}

func (s *Postgres) createTableSQL() string {
	sql := `
CREATE TABLE IF NOT EXISTS %s (
	teamID TEXT NOT NULL,
	clicks NUMERIC,
	UNIQUE(teamID)
);
`
	return fmt.Sprintf(sql, s.tableName)
}

func (s *Postgres) selectAllSQL(limit int) string {
	return fmt.Sprintf("SELECT teamID, clicks FROM %s ORDER BY clicks DESC LIMIT %d", s.tableName, limit)
}

func (s *Postgres) selectOneSQL() string {
	return fmt.Sprintf("SELECT teamID, clicks FROM %s WHERE teamID = $1 LIMIT 1", s.tableName)
}

func (s *Postgres) selectLeaderSQL() string {
	return fmt.Sprintf("SELECT teamID, clicks FROM %s ORDER BY clicks DESC LIMIT 1", s.tableName)
}

func (s *Postgres) upsertSQL() string {
	// We have a specific create operation in the API, so perhaps upserting is a bit bad.
	sql := `
INSERT INTO %s (teamID, clicks) VALUES ($1, $2)
ON CONFLICT (teamID) DO UPDATE SET clicks = %s.clicks + $2
`
	return fmt.Sprintf(sql, s.tableName, s.tableName)
}

// FindByID returns a single team if found by ID.
func (s *Postgres) FindByID(teamID string) (server.Team, error) {

	team := server.Team{}

	rows, err := s.db.Query(s.selectOneSQL(), teamID)
	if err != nil {
		return team, err
	}
	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&team.ID, &team.Clicks)
		if err != nil {
			return team, err
		}
	}
	err = rows.Err()
	if err != nil {
		return team, err
	}

	if team.ID == "" {
		return team, errors.New("team not found")
	}

	return team, nil
}

// CreateTeam creates a new team, an error means the ID is taken.
func (s *Postgres) CreateTeam(teamID string) (server.Team, error) {

	team := server.Team{
		ID: teamID,
	}

	res, err := s.db.Exec(s.upsertSQL(), teamID, 0)
	if err != nil {
		return team, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return team, err
	}
	if rows != 1 {
		return team, errors.New("no rows updated")
	}

	if s.onNewTeam != nil {
		s.onNewTeam <- teamID
	}

	return team, nil
}

// GetLeaderboard returns the highest scoring teams.
func (s *Postgres) GetLeaderboard() (server.Leaderboard, error) {

	leaderboard := server.Leaderboard{}

	// 640 rows ought to be enough for anyone
	rows, err := s.db.Query(s.selectAllSQL(640))
	if err != nil {
		return leaderboard, err
	}
	defer rows.Close()

	team := server.Team{}
	for rows.Next() {
		err := rows.Scan(&team.ID, &team.Clicks)
		if err != nil {
			return leaderboard, err
		}
		leaderboard = append(leaderboard, team)
	}
	err = rows.Err()
	if err != nil {
		return leaderboard, err
	}

	// the database already sorted it for us
	return leaderboard, nil
}

// RecordClicks stores clicks for the given team.
func (s *Postgres) RecordClicks(teamID string, count int64) (server.Team, error) {

	team := server.Team{}

	prevLeader, err := s.findLeader()
	if err != nil {
		log.Printf("no leader found, expected only if no teams played yet")
	}

	res, err := s.db.Exec(s.upsertSQL(), teamID, count)
	if err != nil {
		return team, fmt.Errorf("can't insert team: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return team, fmt.Errorf("can't count affected rows: %w", err)
	}
	if rows < 1 {
		return team, fmt.Errorf("no rows updated: %w", err)
	}

	team, err = s.FindByID(teamID)
	if err != nil {
		return team, fmt.Errorf("can't find updated team: %w", err)
	}

	if s.onNewLeader != nil {
		if prevLeader.Clicks < team.Clicks && prevLeader.ID != teamID {
			s.onNewLeader <- teamID
		}
	}

	return team, nil
}

func (s *Postgres) findLeader() (server.Team, error) {
	leader := server.Team{}

	rows, err := s.db.Query(s.selectLeaderSQL())
	if err != nil {
		return leader, err
	}
	defer rows.Close()

	// the leader query returns at most one row
	for rows.Next() {
		err := rows.Scan(&leader.ID, &leader.Clicks)
		if err != nil {
			return leader, err
		}
	}
	err = rows.Err()
	if err != nil {
		return leader, err
	}

	// if no leader was found we will return a "zero" team
	return leader, nil
}
