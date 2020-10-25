package fleetlock

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// List of ReplyKind
const (
	KindMethodNotAllowed ReplyKind = "method_not_allowed"
	KindMissingHeader    ReplyKind = "missing_header"
	KindDecodeError      ReplyKind = "decode_error"
	KindInternalError    ReplyKind = "internal_error"
	KindLockHeld         ReplyKind = "lock_held"
)

// ReplyKind is used as a Zincati metrics label.
type ReplyKind string

// Reply represents a Fleetlock protocol reply.
type Reply struct {
	// reply identifier
	Kind ReplyKind `json:"kind"`
	// human-friendly reply message
	Value string `json:"value"`
}

// NewReply creates an Reply with a specific kind and a formatted message.
func NewReply(kind ReplyKind, format string, a ...interface{}) Reply {
	return Reply{
		Kind:  kind,
		Value: fmt.Sprintf(format, a...),
	}
}

// encodeReply writes response with the given Reply and HTTP code.
func encodeReply(w http.ResponseWriter, reply Reply) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	switch reply.Kind {
	case KindMethodNotAllowed:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case KindDecodeError, KindMissingHeader:
		w.WriteHeader(http.StatusBadRequest)
	case KindInternalError:
		w.WriteHeader(http.StatusInternalServerError)
	case KindLockHeld:
		w.WriteHeader(http.StatusLocked)
	default:
		w.WriteHeader(http.StatusOK)
	}

	return json.NewEncoder(w).Encode(reply)
}
