# Working with Lostromos

## Suggested Usage

Lostromos is intended to work via a docker image. You will likely want to pass in a config file (via configmap) on
deploy to your Kubernetes cluster, and start the service via `start --config <config_file>`. An example of a basic
deploy can be found [here](../test/data/deploy.yaml). This image is one that we use in integration testing to ensure
that the Lostromos image (not just the binary) is working as expected.