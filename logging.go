package rest

// ----- ---- --- -- -
// Copyright 2019 Oneiro NA, Inc. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"bufio"
	"errors"
	"net"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// LogWriter proxies http.ResponseWriter and logs.
type LogWriter struct {
	http.ResponseWriter
	status int
	length int
}

// Hijack implements http.Hijacker for LogWriter
func (w *LogWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	} else {
		return nil, nil, errors.New("LogWriter's ResponseWriter was not a hijacker")
	}
}

// WriteHeader proxies http.ResponseWriter.WriteHeader
func (w *LogWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// WriteHeader proxies http.ResponseWriter.Write
func (w *LogWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.length += n
	return n, err
}

// LogMW wraps a regular handler and replaces the writer with some logging middleware.
func LogMW(logger log.FieldLogger, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lw := LogWriter{ResponseWriter: w}
		handler.ServeHTTP(&lw, r)
		duration := time.Now().Sub(start)
		logger.WithFields(log.Fields{
			"host":       r.Host,
			"remoteAddr": r.RemoteAddr,
			"method":     r.Method,
			"uri":        r.RequestURI,
			"code":       lw.status,
			"len":        lw.length,
			"ua":         r.Header.Get("User-Agent"),
			"took":       duration,
		}).Info("REQ")
	}
}
