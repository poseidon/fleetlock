# fleetlock

Notable changes between versions.

## Latest

* Build multi-arch container images (amd64, arm64) ([#15](https://github.com/poseidon/fleetlock/pull/15))

## v0.2.0

* Add Prometheus `/metrics` endpoint ([#4](https://github.com/poseidon/fleetlock/pull/4))
* Add JSON error responses ([#9](https://github.com/poseidon/fleetlock/pull/9))
* Fix `-version` command output ([#6](https://github.com/poseidon/fleetlock/pull/6))

## v0.1.0

* Implement the FleetLock protocol backed by Kubernetes coordination API
* Support reboot groups with separate fleetlock-group Leases
* Use a Role with coordination lease create, get, and update
  * Respect `NAMESPACE` if set via downward API, default to "default"

