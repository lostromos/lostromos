# Testing

## Unit Testing

This codebase is expected to be highly tested via unit testing. New functionality should have a unit test if at all
possible, and if not possible that should be explained in the original PR. We may ask you to rewrite the code so that it
is testable, and PRs without tests are more likely to be rejected. Before submitting a PR, run `make test` to ensure no
unit tests have started to fail.

## Code Coverage

Code Coverage is expected to be at least 80%, with a stretch goal of 100%. We track our code coverage via
[codecov.io](https://codecov.io/gh/wpengine/lostromos). To generate a coverage report locally, you can run
`make coverage`.

## Integration Testing

To ensure that Lostromos is working as expected we do integration testing through
[minikube](https://github.com/kubernetes/minikube) to ensure we can watch resources related to our custom resource
definition. We test that

1. We can see resources that already exist.
2. Verify an update has occurred (no processing, just can see it happens).
3. Alert when resources are being deleted.

These tests cover the create/update/delete use cases that lostromos allows. To run that testing in Travis we have to set
up Minikube via this [script](../test/scripts/install_minikube.sh) before actually running the integration tests. It
should be noted that we need minikube 0.22.3 or above due to a bug in 0.22.2 and below which caused issues with kubectl
interactions.

### Testing Docker Image

Integration testing is run against a docker image with kubectl 1.7.x and lostromos installed to ensure that we have a
working lostromos binary. We push the image as a service in minikube to be able to ping the /metrics and /status urls
for use in the tests.

### Running Integration Tests Locally

To run integration tests locally you need to run `minikube start` to get a minikube context, and then run
`make integration-test`. This is assuming you have installed the necessary python requirements via
`make install-python-deps`.