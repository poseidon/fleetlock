package fleetlock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestZincatiID(t *testing.T) {
	cases := []struct {
		machineID string
		expected  string
	}{
		{
			"1c09ca98649c4c7abc779cd04c96812e",
			"978a225b3d7b40e9acd7ce9b62f68444",
		},
		// UUIDs may be dash formatted
		{
			"1c09ca98-649c-4c7a-bc77-9cd04c96812e",
			"978a225b3d7b40e9acd7ce9b62f68444",
		},
	}

	for _, c := range cases {
		actual, err := ZincatiID(c.machineID)
		assert.Nil(t, err)
		assert.Equal(t, c.expected, actual)
	}
}

func TestAppSpecificID(t *testing.T) {
	cases := []struct {
		machineID string
		appID     string
		expected  string
	}{
		// https://docs.rs/libsystemd/0.3.1/src/libsystemd/id128.rs.html#121
		{
			"2e074e9b299c41a59923c51ae16f279b",
			"033b1b9b264441fcaa173e9e5bf35c5a",
			"4d4a86c9c6644a479560ded5d19a30c5",
		},
	}

	for _, c := range cases {
		actual, err := appSpecificID(c.machineID, c.appID)
		assert.Nil(t, err)
		assert.Equal(t, c.expected, actual)
	}
}
