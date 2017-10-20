# Lostrómos

[![Build Status](https://travis-ci.org/wpengine/lostromos.svg?branch=master)](https://travis-ci.org/wpengine/lostromos)
[![codecov](https://codecov.io/gh/wpengine/lostromos/branch/master/graph/badge.svg)](https://codecov.io/gh/wpengine/lostromos)

**Under active development. Not ready for production usage.**

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

## Getting Started

- [Continuous Integration](./docs/travisCI.md)
- [Development](./docs/development.md)
