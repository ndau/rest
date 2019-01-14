package main

import (
	"net/http"
	"strconv"

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
