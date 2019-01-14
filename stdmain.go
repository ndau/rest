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

func logAndDie(logger *log.Entry, reason string) func() {
	return func() {
		logger.Info("shutting down because of " + reason)
		os.Exit(0)
	}
}

// stuff I want to do:
// [x] generic signal handler
// [x] included cors handling
// [x] print api docs to arbitrary file (instead of just stdout)
// [x] add a standard /prefix/docs that returns documentation for this part of the API
// [ ] figure out a way to generate unified docs through a /docs endpoint that knows about the rest
// [x] either strip prefix or inject one at startup time
// [ ] request throttling
// [x] honeycomb logging
// [x] unified config handling

// Builder is the interface to which all service builders must conform
type Builder interface {
	Build(logger *log.Entry, path string) *boneful.Service
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
	cf.AddString("HONEYCOMB_DATASET", "ndev_backend")
	cf.AddString("HONEYCOMB_KEY", "")
	return cf
}

// StandardMain is what should be called to start the service.
// It never returns.
func StandardMain(cf *Config, builder Builder) {
	cf.Load()

	docs := cf.GetString("docs")
	if docs != "" {
		var outf = os.Stdout
		var err error
		if docs != "-" {
			outf, err = os.Create(docs)
			if err != nil {
				log.Fatalf("could not write docs to %s", docs)
			}
		}
		svc := builder.Build(nil, cf.GetString("rootpath"))
		svc.GenerateDocumentation(outf)
		return
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
		Addr:    fmt.Sprintf(":%v", cf.GetInt("port")),
		Handler: handler,
	}
	logger.WithField("port", cf.GetInt("port")).Info("server listening")

	// if your service requires cleanup before exiting, or can be restarted or
	// reloaded, set up the appropriate functions and pass them here.
	WatchSignals(nil, logAndDie(logger, "SIGINT"), logAndDie(logger, "SIGTERM"))
	log.Fatal(server.ListenAndServe())
}
