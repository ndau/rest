package main

// ----- ---- --- -- -
// Copyright 2019, 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----


import (
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-zoo/bone"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/reqres"
)

// Count returns a HandlerFunc that counts from start to end
func Count() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		first, err := strconv.Atoi(bone.GetValue(r, "first"))
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("'first' parameter did not parse as an integer", http.StatusBadRequest))
			return
		}
		last, err := strconv.Atoi(bone.GetValue(r, "last"))
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("'last' parameter did not parse as an integer", http.StatusBadRequest))
			return
		}
		if first > last {
			reqres.RespondJSON(w, reqres.NewAPIError("'first' must be less than 'last'", http.StatusBadRequest))
			return
		}
		if first+100 < last {
			reqres.RespondJSON(w, reqres.NewAPIError("cannot return more than 100 values", http.StatusBadRequest))
			return
		}

		resp := make([]int, last-first+1)
		for i := first; i <= last; i++ {
			resp[i-first] = i
		}
		reqres.RespondJSON(w, reqres.OKResponse(resp))
	}
}

// Passthrough passes the query onto a child server (which
// is expected to be the same server)
func Passthrough(u string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		first := bone.GetValue(r, "first")
		last := bone.GetValue(r, "last")

		client := http.Client{
			Timeout: 1 * time.Second,
		}
		req, _ := http.NewRequest("GET", u+"/count/"+first+"/"+last, nil)
		resp, err := client.Do(req)
		if err != nil {
			reqres.RespondJSON(w, reqres.NewAPIError("bad response from passthrough", http.StatusInternalServerError))
			return
		}

		body, _ := ioutil.ReadAll(resp.Body)
		realresp := reqres.Response{
			Bd:  body,
			Sts: resp.StatusCode,
		}
		reqres.RespondJSON(w, realresp)
	}
}

// Die kills the server after a 1 second delay
func Die() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := bone.GetValue(r, "code")
		exitcode, _ := strconv.Atoi(code)
		go func() {
			time.Sleep(1 * time.Second)
			os.Exit(exitcode)
		}()
		reqres.RespondJSON(w, reqres.OKResponse("Shutting down in 1 sec"))
	}
}
