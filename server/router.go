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
	_ "embed" // needed to go:embed the API yaml
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// Route maps a path pattern to a handler function.
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

// Routes is a set of routes.
type Routes []Route

// NewRouter creates a router with routes for the MMOCG API
func NewRouter() *mux.Router {
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range apiRoutes {
		var handler http.Handler
		handler = route.HandlerFunc
		handler = Logger(handler, route.Name)

		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	return router
}

//go:embed openapi.yaml
var swaggerYaml string

// APIDefinition serves the swagger description for this service
func APIDefinition(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(swaggerYaml))
}

var apiRoutes = Routes{
	Route{
		"Index",
		"GET",
		"/v1",
		APIDefinition,
	},

	Route{
		"Click",
		strings.ToUpper("Post"),
		"/v1/team/{teamId}/click",
		Click,
	},

	Route{
		"GetLeaderboard",
		strings.ToUpper("Get"),
		"/v1/leaderboard",
		GetLeaderboard,
	},

	Route{
		"GetTeamById",
		strings.ToUpper("Get"),
		"/v1/team/{teamId}",
		GetTeamByID,
	},

	Route{
		"UpdateTeam",
		strings.ToUpper("Post"),
		"/v1/team/{teamId}",
		UpdateTeam,
	},
}
