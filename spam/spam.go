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

// Package spam spams "important" announcements.
package spam

import (
	"bytes"
	"html/template"
	"log"

	psacfg "github.com/fabjan/psa/configure"
	"go.uber.org/ratelimit"
)

func renderAnnouncement(tmpl *template.Template, msg string) string {
	var buf bytes.Buffer
	tmpl.Execute(&buf, struct{ Message string }{
		Message: msg,
	})
	return template.HTMLEscapeString(buf.String())
}

// Handler listens for team updates and spams announcements.
type Handler struct {
	onNewTeam, onNewLeader chan string
	cfg                    psacfg.AppConfig
	spamPerSecond          int
}

// NewHandler creates a new spam handler.
func NewHandler(onNewTeam, onNewLeader chan string) *Handler {
	cfg, err := psacfg.FromEnv()
	if err != nil {
		log.Fatalf("failed announcement config: %v", err)
	}
	return &Handler{
		spamPerSecond: 1, // TODO configurable?
		onNewTeam:     onNewTeam,
		onNewLeader:   onNewLeader,
		cfg:           cfg,
	}
}

// Go starts the channel listener/spamming forever loop.
func (h *Handler) Go() {
	rl := ratelimit.New(h.spamPerSecond)
	announcers := h.cfg.Announcers()
	tmpl := h.cfg.MessageTemplate
	for {
		rl.Take()
		// TODO a quit channel for graceful shutdown
		select {
		case challenger := <-h.onNewTeam:
			msg := renderAnnouncement(tmpl, "A challenger appears! ("+challenger+")")
			for _, a := range announcers {
				err := a.Announce(msg)
				if err != nil {
					log.Printf("failed to send announcement: %v", err)
				}
			}
		case leader := <-h.onNewLeader:
			msg := renderAnnouncement(tmpl, leader+" is in now the lead!")
			for _, a := range announcers {
				err := a.Announce(msg)
				if err != nil {
					log.Printf("failed to send announcement: %v", err)
				}
			}
		}

	}
}
