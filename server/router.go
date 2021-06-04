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
