package fleetlock

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ReplyKind is a reply type identifier.
//
// It is designed to be tracked in Zincati metrics.
// Fleetlock specification states that this MUST have a bounded/small cardinality.
type ReplyKind string

// List of ReplyKind
const (
	ErrorMethodNotAllowed ReplyKind = "method_not_allowed"
	ErrorMissingHeader    ReplyKind = "missing_header"
	ErrorDecodingRequest  ReplyKind = "failed_decoding_request"
	ErrorInternal         ReplyKind = "internal_error"
	ErrorLocked           ReplyKind = "locked"
)

// Reply holds data used for fleetlock replies (currently only used for error replies).
type Reply struct {
	// reply type identifier
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

// encodeReply writes a reply with the specified Reply and inferred http status.
func encodeReply(w http.ResponseWriter, fleetlockErr Reply) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	switch fleetlockErr.Kind {
	case ErrorMethodNotAllowed:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case ErrorDecodingRequest, ErrorMissingHeader:
		w.WriteHeader(http.StatusBadRequest)
	case ErrorInternal:
		w.WriteHeader(http.StatusInternalServerError)
	case ErrorLocked:
		w.WriteHeader(http.StatusLocked)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	json.NewEncoder(w).Encode(fleetlockErr)
}
