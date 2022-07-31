# fleetlock [![GoDoc](https://pkg.go.dev/badge/github.com/poseidon/fleetlock.svg)](https://pkg.go.dev/github.com/poseidon/fleetlock) [![Quay](https://img.shields.io/badge/container-quay-green)](https://quay.io/repository/poseidon/fleetlock) [![Workflow](https://github.com/poseidon/fleetlock/actions/workflows/test.yaml/badge.svg)](https://github.com/poseidon/fleetlock/actions/workflows/test.yaml?query=branch%3Amain) [![Sponsors](https://img.shields.io/github/sponsors/poseidon?logo=github)](https://github.com/sponsors/poseidon) [![Twitter](https://img.shields.io/badge/follow-news-1da1f2?logo=twitter)](https://twitter.com/poseidonlabs)

`fleetlock` is a reboot coordinator for Fedora CoreOS nodes in Kubernetes clusters. It implements the [FleetLock](https://coreos.github.io/zincati/development/fleetlock/protocol/) protocol for use as a [Zincati](https://github.com/coreos/zincati) lock [strategy](https://github.com/coreos/zincati/blob/master/docs/usage/updates-strategy.md) backend.

## Usage

Zincati runs on-host (`zincati.service`). Declare a Zincati `fleet_lock` strategy when provisioning Fedora CoreOS nodes. Set `base_url` for host nodes to access the in-cluster `fleetlock` Service (e.g. known ClusterIP).

```yaml
variant: fcos
version: 1.4.0
storage:
  files:
    - path: /etc/zincati/config.d/55-update-strategy.toml
      contents:
        inline: |
          [updates]
          strategy = "fleet_lock"
          [updates.fleet_lock]
          base_url = "http://10.3.0.15/"
```

Apply the `fleetlock` Deployment, Service (with ClusterIP), and ServiceAccount.

```
kubectl apply -f examples/k8s
```

Inspect the fleetlock Lease object.

```
$ kubectl get leases -n default
NAME                HOLDER                             AGE
fleetlock-default   049ad0f57ade4723a48692b7b692c318   4m50s
```

### Configuration

Configure the server via flags.

| flag       | description  | default      |
|------------|--------------|--------------|
| -address   | HTTP listen address | 0.0.0.0:8080 |
| -log-level | Logger level | info |
| -version   | Show version | NA   |
| -help      | Show help    | NA   |

Or via environment variables.

| variable   | description            | default   |
|------------|------------------------|-----------|
| NAMESPACE  | Kubernetes Namespace   | "default" |
| KUBECONFIG | Development Kubeconfig | NA        |

### Typhoon

For Typhoon clusters, add the Zincati config a [snippet](https://typhoon.psdn.io/advanced/customization/#fedora-coreos).

```tf
module "nemo" {
  ...
  controller_snippets = [
    file("./snippets/zincati-strategy.yaml"),
  ]
  worker_snippets = [
    file("./snippets/zincati-strategy.yaml"),
  ]
}
```

## Manual Intervention

`fleetlock` coordinates OS auto-updates to avoid concurrent node updates or a potential bad auto-update continuing. Zincati obtains a reboot lease lock before finalization (i.e reboot).

If an auto-update fails, the lease continues to be held by design. An admin should investigate the node failure and decide whether it is safe to remove the lease.

```
$ kubectl get leases
$ kubectl delete lease fleetlock-default
```

## Metrics

`fleetlock` serves Prometheus `/metrics` from Go, process, and custom collectors.

| name                 | description                                         |
|----------------------|-----------------------------------------------------|
| fleetlock_lock_state | State of the fleetlock lease (0 unlocked, 1 locked) |
| fleetlock_lock_transition_count | Number of fleetlock lease transitions    |
| fleetlock_lock_request_count   | Number of lock requests   |
| fleetlock_unlock_request_count | Number of unlock requests |

## Development

To develop locally, build and run the executable.

### Static Binary

Build the static binary.

```
make build
```

### Container Image

Build the container image.

```
make image
```

### Run

Run the executable.

```
export KUBECONFIG=some-dev-kubeconfig
./bin/fleetlock
```

Use curl to emulate a Zincati FleetLock client.

```json
{
  "client_params": {
    "id": "c988d2509fdf5cdcbed39037c56406fb",
    "group": "default"
  }
}
```

Request a reboot lock.


```
curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/pre-reboot
```

Release a reboot lock.

```
curl -H "fleet-lock-protocol: true" -d @examples/body.json http://127.0.0.1:8080/v1/steady-state
```

## Related

* [Zincati Guide](https://docs.fedoraproject.org/en-US/fedora-coreos/auto-updates/)
* [Zincati Docs](https://github.com/coreos/zincati/blob/master/docs/usage/updates-strategy.md)
* [FleetLock Protocol](https://coreos.github.io/zincati/development/fleetlock/protocol/)
