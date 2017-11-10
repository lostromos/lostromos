# Lostrómos

[![Build Status](https://travis-ci.org/wpengine/lostromos.svg?branch=master)](https://travis-ci.org/wpengine/lostromos)
[![codecov](https://codecov.io/gh/wpengine/lostromos/branch/master/graph/badge.svg)](https://codecov.io/gh/wpengine/lostromos)

**NOTE**: Under active development. Not ready for production usage.

Lostromos is a service that creates Kubernetes resources based on a Custom Resource
endpoint in the Kubernetes API. It is an implementation of the [Operator
pattern](https://coreos.com/blog/introducing-operators.html) established by CoreOS.

It creates resources using Go templates, using each Custom Resource as the values
to use during templating. It also watches the CR endpoint for creates, updates,
and deletes and reconciles the corresponding k8s resources.

It is intended to be deployed into a Kubernetes cluster. Its main configuration
details are:

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
setup.

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
make LOSTROMOS_IP_AND_PORT=`minikube service lostromos --url | cut -c 8-` integration-tests
eval $(minikube docker-env -u) # Unlinks minikube and docker.
```

## Getting Started

- [Development](./docs/development.md)
- [Continuous Integration](./docs/continuousIntegration.md)
  - [Testing](./docs/testing.md)
- [Deployment](./docs/deployment.md)

## Contributing

See [Contributing](./CONTRIBUTING.md) to get started.

## Report a Bug

To report an issue or suggest an improvement please open an [issue](https://github.com/wpengine/lostromos/issues).