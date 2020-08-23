package fleetlock

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Message represents a FleetLock protocol client request.
type Message struct {
	ClientParmas struct {
		ID    string `json:"id"`
		Group string `json:"group"`
	} `json:"client_params"`
}

// decodeMessage decodes a Message from a request.
func decodeMessage(w http.ResponseWriter, req *http.Request) (*Message, error) {
	msg := &Message{}
	err := json.NewDecoder(req.Body).Decode(msg)
	if err != nil {
		return nil, err
	}

	if msg.ClientParmas.ID == "" {
		return nil, fmt.Errorf("message missing id: %v", msg)
	}

	if msg.ClientParmas.Group == "" {
		return nil, fmt.Errorf("message missing group: %v", msg)
	}

	return msg, nil
}
