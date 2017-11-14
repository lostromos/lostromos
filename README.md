# Lostrómos

[![Build Status](https://travis-ci.org/wpengine/lostromos.svg?branch=master)](https://travis-ci.org/wpengine/lostromos)
[![codecov](https://codecov.io/gh/wpengine/lostromos/branch/master/graph/badge.svg)](https://codecov.io/gh/wpengine/lostromos)

**NOTE**: Under active development. Not ready for production usage.

## Problem statement

Lostrómos is a way to manage deployments through Custom Resources. This allows a user to harness the power and
flexibility of the Kubernetes platform to more easily control resources through outside applications.

WP Engine uses Lostrómos to customize deployments into GKE. As we spin up new clusters, we use another tool to monitor
the google api for changes and update Custom Resources as they happen. Lostrómos watches for Custom Resources and
applies a predefined template based on the information received. This allows us to deconstruct a larger service by
deliniating based on functionality. Our applications now manage resources, and Lostrómos handles the deployments.

## How it works

Lostrómos is a service that creates Kubernetes resources based on a Custom Resource endpoint in the Kubernetes API. It
is an implementation of the [Operator pattern](https://coreos.com/blog/introducing-operators.html) established by
CoreOS.

Lostrómos manages objects (pods, sets, jobs - anything that can be managed with a CRD) via Helm or Go templates, using
values from Custom Resources as new events are captured from the Kubernetes API. It applies templates to the
corresponding objects to reconcile with Kubernetes.

It is intended to be deployed into a Kubernetes cluster. Its main configuration details are:

- An API endpoint of a Custom Resource Definition to watch
- A set of go templates to apply for each Custom Resource

Its configuration could also include shared values to use in the templating (eg.
docker image in deployments, a common annotation or label).

## Dependencies

| Dependency | Version |
| ---------- | ------- |
| `Golang` | 1.9.0+ |
| `Minikube` | 0.22.3+ |
| `Docker` | 17.09.0+ |
| `Python` | 3.0+ |

## Quick Start

**NOTE**: This assumes you have all of the above dependency requirements.

Run the following script (changing out os_version for darwin/linux/windows depending on your system) to get a basic
setup. This script will install Go and Python dependencies, build Lostrómos, build a docker image with Lostrómos, then
run it in Minikube and perform integration testing.

```bash
make install-go-deps
make vendor
make install-python-deps
make build-cross
./out/lostromos-os_version-amd64 version
minikube start
eval $(minikube docker-env) # This links docker with minikube so that the image you build in the next step will be available.
make docker-build-test
kubectl create -f test/data/crd.yml
kubectl expose pod lostromos --type=LoadBalancer
make LOSTROMOS_IP_AND_PORT=`minikube service lostromos --url | cut -c 8-` integration-tests
eval $(minikube docker-env -u) # Unlinks minikube and docker.
```

## Using Lostrómos

- [Working with Lostrómos](./docs/workingWithLostromos.md)
- [Development](./docs/development.md)
- [Continuous Integration](./docs/continuousIntegration.md)
  - [Testing](./docs/testing.md)
- [Deployment](./docs/deployment.md)

## Contributing

See [Contribution Guildelines](./CONTRIBUTING.md) to get started.

## Report a Bug

To report an issue or suggest an improvement please open an [issue](https://github.com/wpengine/lostromos/issues).