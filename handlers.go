package fleetlock

import (
	"fmt"
	"net/http"
)

const (
	fleetLockHeaderKey = "fleet-lock-protocol"
)

// POSTHandler returns a handler that requires the POST method.
func POSTHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			http.Error(w, "required method POST", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}

// HeaderHandler returns a handler that requires a given header key/value.
func HeaderHandler(key, value string, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get(key) != value {
			errmsg := fmt.Sprintf("missing required header %s: %s", key, value)
			http.Error(w, errmsg, http.StatusBadRequest)
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}
