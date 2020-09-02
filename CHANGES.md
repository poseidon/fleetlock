# fleetlock

Notable changes between versions.

## Latest

* Add Prometheus `/metrics` endpoint ([#4](https://github.com/poseidon/fleetlock/pull/4))

## v0.1.0

* Implement the FleetLock protocol backed by Kubernetes coordination API
* Support reboot groups with separate fleetlock-group Leases
* Use a Role with coordination lease create, get, and update
  * Respect `NAMESPACE` if set via downward API, default to "default"

