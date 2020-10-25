package fleetlock

import (
	"net/http"
)

const (
	fleetLockHeaderKey = "fleet-lock-protocol"
)

// POSTHandler returns a handler that requires the POST method.
func POSTHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			encodeReply(w, NewReply(KindMethodNotAllowed, "required method POST"))
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
			encodeReply(w, NewReply(KindMissingHeader, "missing required header %s: %s", key, value))
			return
		}
		next.ServeHTTP(w, req)
	}
	return http.HandlerFunc(fn)
}
