![Lostrómos logo](docs/images/logo.png)

[![Build Status](https://travis-ci.org/lostromos/lostromos.svg?branch=master)](https://travis-ci.org/lostromos/lostromos)
[![codecov](https://codecov.io/gh/lostromos/lostromos/branch/master/graph/badge.svg)](https://codecov.io/gh/lostromos/lostromos)
[![Go Report Card](https://goreportcard.com/badge/github.com/lostromos/lostromos)](https://goreportcard.com/report/github.com/lostromos/lostromos)

# Lostrómos

Lostrómos is a templating [operator](https://coreos.com/blog/introducing-operators.html).

*Please note that the documentation on Kubernetes Operators is somewhat out of
date. [Third party resources](https://kubernetes.io/docs/tasks/access-kubernetes-api/extend-api-third-party-resource/)
have been deprecated, and operators now watch [Custom Resources](https://kubernetes.io/docs/concepts/api-extension/custom-resources/).*

## Table of Contents

* [Overview](#overview)
  * [Problem Statement](#problem-statement)
  * [How It Works](#how-it-works)
  * [Use Cases](#use-cases)
* [Using Lostrómos](docs/usinglostromos.md#usinglostromos)
  * [Recommended Reading](docs/usinglostromos.md#reading)
  * [Quick Start](docs/usinglostromos.md#quickstart)
  * [Tutorial](docs/usinglostromos.md#tutorial)
  * [Customization](docs/usinglostromos.md#customization)
    * [Events](docs/events.md)
    * [Using Helm](docs/helm.md)
  * [Deploying Lostrómos](docs/usinglostromos.md#deployment)
  * [Logs](docs/usinglostromos.md#logs)
* [Contributing to Lostrómos](docs/development.md#contributing)
  * [Development](docs/development.md#development)
  * [Testing](docs/development.md#testing)
  * [Continuous Integration](docs/development.md#ci)
  * [Contribution Guidelines](CONTRIBUTING.md)
    * [Report a Bug](CONTRIBUTING.md#bugs)

## Overview

### Problem Statement

Managing, sharing, and controlling an application's operational domain knowledge
can be prone to human error and may create points of failure. Instead of
maintaining lists, databases, and/or logic structures to control this
information, Lostrómos automates maintenance of this information with only the
need for a predefined template.

### How It Works

Lostrómos is a Kubernetes operator. It watches a Custom Resource (CR) endpoint.
When a change is detected, it uses the information in the CR to fill a
template. This template is applied either via
[kubectl](https://kubernetes.io/docs/user-guide/kubectl-overview/) or
[Helm](https://docs.helm.sh/).

### Use Cases

*Control access to creation of Kubernetes resources*

As a Kubernetes admin, allow developers to create instances of an application
for development purposes without giving them direct access to deploy to the
production cluster. Developers can create a CR, and with Lostrómos, the instance
is deployed with development specific operational parameters (such as a test
database or a specific application package).

*Automate deployment of services alongside your application*

Deploy a Kubernetes application and an accompanying monitoring service that
relies on operational data from that application (such as an IP address) by
creating a single CR.

*Eliminate maintenance of application operational knowledge for deployments*

WP Engine previously used Lostrómos in conjunction with another tool to customize VM
deployments into GCE. Each VM instance offloaded some of its workload to a
separate Kubernetes application. As new VMs were created in GCE, this other tool
monitored the Google API for these changes and created a CR as they happened.
Lostrómos watched for changes to this CR endpoint and created a Helm release
by combining information from the new CR and a predefined template. This allowed
WP Engine to deconstruct some of the work for deployment into GCE and reduce
maintenance work around sharing the data between applications.
