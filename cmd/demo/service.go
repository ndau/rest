package main

import (
	"github.com/kentquirk/boneful"
	log "github.com/sirupsen/logrus"
)

// JSON is the MIME type that we process
const JSON = "application/json"

type countService struct{}

// Build builds the service from the set of routes as defined
// Path is the top-level path that gets you to this service
func (c *countService) Build(logger *log.Entry, path string) *boneful.Service {
	svc := new(boneful.Service).
		Path(path).
		Doc(`This provides the API for the sample server.
		`)

	svc.Route(svc.GET("/count/:first/:last").To(Count()).
		Doc("Returns an array of numbers from first to last.").
		Notes("Just a dummy endpoint to show some techniques").
		Operation("Count").
		Produces(JSON).
		Writes([]int{4, 5, 6}))

	return svc
}
