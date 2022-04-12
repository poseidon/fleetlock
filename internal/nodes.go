package fleetlock

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"strings"

	"encoding/hex"
)

// Zincati requests include a node (i.e. agent) UUID identifier, assigned to
// match `systemd-id128 machine-id -a APP_ID` for Zincati's chosen App ID.
// systemd and rust's libsystemd assign IDs as the SHA256 HMAC of the systemd
// machine-id and the app-id.
//
// Kubelet reports the systemd machine-id as the MachineID (if /etc/machine-id
// is mounted) or as the System UUID (via sysfs /sys/class/dmi/id/product_uuid)
// on most platforms (except Azure or bare-metal). Favor the explicit mount.
//
// We can compute the app-specific ID as systemd or libsystemd would, in
// order to map Zincati node IDs to Kubernetes nodes. This is valuable to
// provide better logs or drain nodes in advance of reboots.
//
// Important: Implementations aim to match systemd (C) which encodes a 128 bit
// UUID into a 16 byte array. In modern languages, be sure to use the 16 byte
// hex encoding, not the 32 byte string representation to get the correct hmac.
//
// Related
// - https://github.com/coreos/zincati/pull/4
// - https://docs.rs/libsystemd/0.3.1/src/libsystemd/id128.rs.html#38-57
// - https://github.com/systemd/systemd/blob/011d129cf42a780c30a525f0ad00985422e25fdb/src/libsystemd/sd-id128/sd-id128.c

const (
	// Zincati App ID
	// https://github.com/lucab/zincati/blob/17d5e2adf13ee9a98cebc662735a2084949e589b/src/identity/mod.rs#L9
	zincatiAppID = "de35106b6ec24688b63afddaa156679b"
)

// ZincatiID computes the Zincati node ID for a systemd machine ID.
func ZincatiID(machineID string) (string, error) {
	return appSpecificID(machineID, zincatiAppID)
}

// appSpecificID computes a systemd-style app-specific identifier given a
// machine ID and an application ID.
func appSpecificID(machineID string, appID string) (string, error) {
	// Remove any UUID dash formatting
	machineID = strings.ReplaceAll(machineID, "-", "")

	machineBytes, err := hex.DecodeString(machineID)
	if err != nil {
		return "", err
	}
	appBytes, err := hex.DecodeString(appID)
	if err != nil {
		return "", err
	}

	// NOT for security uses
	mac := hmac.New(sha256.New, machineBytes)
	mac.Write(appBytes)
	sum := mac.Sum(nil)

	// UUID v4 settings
	// https://docs.rs/libsystemd/0.3.1/src/libsystemd/id128.rs.html#52-54
	// https://github.com/systemd/systemd/blob/5a7eb46c0206411d380543021291b4bca0b6f59f/src/libsystemd/sd-id128/id128-util.c#L199
	sum[6] = (sum[6] & 0x0F) | 0x40
	sum[8] = (sum[8] & 0x3F) | 0x80

	id := string(sum)[:16]
	return fmt.Sprintf("%x", id), nil
}
