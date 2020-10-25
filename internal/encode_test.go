package fleetlock

import (
	"fmt"
	"io/ioutil"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeReply(t *testing.T) {
	cases := []struct {
		reply            Reply
		expectedStatus   int
		expectedResponse string
	}{
		{
			reply:            NewReply(KindMethodNotAllowed, "required method POST"),
			expectedStatus:   405,
			expectedResponse: `{"kind": "method_not_allowed", "value": "required method POST"}`,
		},
		{
			reply:            NewReply(KindMissingHeader, "missing required header %s: %s", fleetLockHeaderKey, "true"),
			expectedStatus:   400,
			expectedResponse: `{"kind": "missing_header", "value": "missing required header fleet-lock-protocol: true"}`,
		},
		{
			reply:            NewReply(KindDecodeError, "error decoding message"),
			expectedStatus:   400,
			expectedResponse: `{"kind": "decode_error", "value": "error decoding message"}`,
		},
		{
			reply:            NewReply(KindInternalError, "error getting reboot lease"),
			expectedStatus:   500,
			expectedResponse: `{"kind": "internal_error", "value": "error getting reboot lease"}`,
		},
		{
			reply:            NewReply(KindLockHeld, "reboot lease unavailable, held by %s", "e0f3745b108f471cbd4883c6fbed8cdd"),
			expectedStatus:   423,
			expectedResponse: `{"kind": "lock_held", "value": "reboot lease unavailable, held by e0f3745b108f471cbd4883c6fbed8cdd"}`,
		},
		{
			reply:            NewReply("other", "message"),
			expectedStatus:   200,
			expectedResponse: `{"kind": "other", "value": "message"}`,
		},
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%v-%v", c.reply.Kind, c.reply.Value), func(t *testing.T) {
			w := httptest.NewRecorder()
			encodeReply(w, c.reply)

			assert.Equal(t, c.expectedStatus, w.Code, "Expected status code %v", c.expectedStatus)

			body, _ := ioutil.ReadAll(w.Body)
			assert.JSONEq(t, c.expectedResponse, string(body), "Unexpected JSON output in response")
		})
	}
}
