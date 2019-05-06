package rest

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kentquirk/boneful"
	"github.com/oneiro-ndev/o11y/pkg/honeycomb"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

// WatchSignals registers with the operating system to receive
// specific signals and can call functions on those signals.
// In the case of SIGTERM, if the function returns, os.Exit is
// called with a normal exit code of 0.
func WatchSignals(fhup, fint, fterm func()) {
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM)
		for {
			sig := <-sigchan
			switch sig {
			case syscall.SIGHUP:
				if fhup != nil {
					fhup()
				}
			case syscall.SIGINT:
				if fint != nil {
					fint()
				}
			case syscall.SIGTERM:
				if fterm != nil {
					fterm()
				}
				os.Exit(0)
			}
		}
	}()
}

// func (c *Cycle) interceptSignals() {
// 	ch := make(chan os.Signal, 1)
// 	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
// 	for {
// 		select {
// 		case sig := <-ch:
// 			log.Printf("lifecycle received %q, terminating\n", sig)
// 			c.End()
// 			return
// 		case <-c.Ctx.Done():
// 			return
// 		}
// 	}

// }

// FatalFunc returns a function that logs and shuts down the service
func FatalFunc(builder Builder, reason string) func() {
	return func() {
		builder.GetLogger().Fatalf("shutting down because of " + reason)
	}
}

// stuff I want to do:
// [x] generic signal handler
// [x] included cors handling
// [x] print api docs to arbitrary file (instead of just stdout)
// [x] add a standard /prefix/docs that returns documentation for this part of the API
// [x] either strip prefix or inject one at startup time
// [x] honeycomb logging
// [x] unified config handling
// [x] super simple setup and use
// [x] default health check
// [ ] simple dockerfiles for build and deploy
// [ ] circleci sample for easy deploy on merge/tag
// [ ] set up AWS ALB routing and AWS ECS for zero-downtime deploys
// [ ] wrapper for AWS Dynamo for easy data storage
// [ ] figure out a way to generate unified docs through a /docs endpoint that knows about the rest
// [ ] auth token middleware for APIs
// [ ] request throttling to limit bandwidth
// [ ] contexts for cancellation

// Builder is the interface to which all service builders must conform
type Builder interface {
	Build(logger *log.Entry, path string) *boneful.Service
	GetLogger() *log.Entry
}

// DefaultConfig creates a default configuration including the values
// that are used by the standard server.
func DefaultConfig() *Config {
	cf := NewConfig()
	cf.AddString("docs", "")
	cf.AddStringArray("CORS_ORIGINS", "*")
	cf.AddStringArray("CORS_METHODS", "GET", "POST", "PUT", "DELETE")
	cf.AddFlag("CORS_DEBUG", false)
	cf.AddInt("port", 8080)
	cf.AddString("rootpath", "/")
	cf.AddDuration("READ_TIMEOUT", "5s")
	cf.AddDuration("WRITE_TIMEOUT", "5s")
	cf.AddString("HONEYCOMB_DATASET", "ndev_backend")
	cf.AddString("HONEYCOMB_KEY", "")
	return cf
}

// StandardSetup is what should be called to set up the service before
// running it. It returns a server, or possibly nil.
func StandardSetup(cf *Config, builder Builder) *http.Server {
	docs := cf.GetString("docs")
	if docs != "" {
		var outf = os.Stdout
		var err error
		if docs != "-" {
			outf, err = os.Create(docs)
			if err != nil {
				log.Fatalf("could not write docs to %s", docs)
			}
			return nil
		}
		svc := builder.Build(nil, cf.GetString("rootpath"))
		svc.GenerateDocumentation(outf)
		return nil
	}

	// create the logger
	hlog := honeycomb.Setup(log.New())
	logger := hlog.WithFields(log.Fields{
		"rootpath": cf.GetString("rootpath"),
	})
	// now create the service
	svc := builder.Build(logger, cf.GetString("rootpath"))
	// wrap it in logging middleware
	logmux := LogMW(logger, svc.Mux())
	// and then in cors
	c := cors.New(cors.Options{
		// allow * by default
		// in production we may want to be more picky, depending on whether
		// we want to allow third parties to access this api from apps
		// that we don't control.
		AllowedOrigins: cf.GetStringArray("CORS_ORIGINS"),
		// These are the basic REST methods; update if needed
		AllowedMethods: cf.GetStringArray("CORS_METHODS"),
		// You can turn this on for debugging
		Debug: cf.GetFlag("CORS_DEBUG"),
		// We don't currently need/use credentials. But that can change.
		AllowCredentials: false,
	})
	handler := c.Handler(logmux)

	// now create the server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%v", cf.GetInt("port")),
		Handler:      handler,
		ReadTimeout:  cf.GetDuration("READ_TIMEOUT"),
		WriteTimeout: cf.GetDuration("WRITE_TIMEOUT"),
	}
	logger.WithField("port", cf.GetInt("port")).Info("server listening")
	return server
}
