package main

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"github.com/kentquirk/boneful"
	"github.com/oneiro-ndev/rest"
	log "github.com/sirupsen/logrus"
)

// JSON is the MIME type that we process
const JSON = "application/json"

type countService struct {
	Logger         *log.Entry
	Svc            *boneful.Service
	PassthroughURL string
}

// verify that it conforms to Builder
var _ rest.Builder = (*countService)(nil)

// Logger implements part of the Builder interface
func (c *countService) GetLogger() *log.Entry {
	return c.Logger
}

// Build builds the service from the set of routes as defined
// Path is the top-level path that gets you to this service
func (c *countService) Build(logger *log.Entry, path string) *boneful.Service {
	c.Logger = logger

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

	svc.Route(svc.GET("/die/:code").To(Die()).
		Doc("Kills the server with the given exit code").
		Operation("Die").
		Produces(JSON).
		Writes("dying"))

	svc.Route(svc.GET("/passthrough/:first/:last").
		To(Passthrough(c.PassthroughURL)).
		Doc("Passes the count query on to the child service.").
		Notes("Another dummy endpoint to show some techniques").
		Operation("Passthrough").
		Produces(JSON).
		Writes([]int{4, 5, 6}))

	return svc
}
