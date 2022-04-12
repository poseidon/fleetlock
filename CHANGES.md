# fleetlock

Notable changes between versions.

## Latest

* Identify Kubelet nodes via `MachineID` instead of `SystemUUID` (**action required**) ([#96](https://github.com/poseidon/fleetlock/pull/96))
  * Host `SystemUUID` and `MachineID` are identical on most platforms, but not on Azure or bare-metal
  * Kubelet detected `SystemUUID` via sysfs, but `MachineID` requires Kubelet
  be able to read `/etc/machine-id` from the host (`ro` mount when Kubelet
  run as a container) (**action required**)
  * Add support for Azure and bare-metal plaforms by changing to `MachineID`
* Reset `fleet_lock_state` gauge on first `lock` or `unlock` call ([#71](https://github.com/poseidon/fleetlock/pull/71))
* Improve reject reply when a client attempts to unlock a lock it doesn't own ([#71](https://github.com/poseidon/fleetlock/pull/71))

## v0.3.0

* Add support for Kubernetes node draining ([#51](https://github.com/poseidon/fleetlock/pull/51))
* Automate base image, Go version, and module dependency updates
* Build multi-arch container images (amd64, arm64) ([#15](https://github.com/poseidon/fleetlock/pull/15))
* Switch to using the Prometheus `collectors` package ([#37](https://github.com/poseidon/fleetlock/pull/37))

## v0.2.0

* Add Prometheus `/metrics` endpoint ([#4](https://github.com/poseidon/fleetlock/pull/4))
* Add JSON error responses ([#9](https://github.com/poseidon/fleetlock/pull/9))
* Fix `-version` command output ([#6](https://github.com/poseidon/fleetlock/pull/6))

## v0.1.0

* Implement the FleetLock protocol backed by Kubernetes coordination API
* Support reboot groups with separate fleetlock-group Leases
* Use a Role with coordination lease create, get, and update
  * Respect `NAMESPACE` if set via downward API, default to "default"

