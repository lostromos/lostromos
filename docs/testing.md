# Testing

## Unit Testing

This codebase is expected to be highly tested via unit testing. New functionality should have a unit test if at all
possible, and if not possible that should be explained in the original PR if nothing else. Before doing a PR, it's
suggested you at least run `make test` to ensure no unit tests have started to fail.

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

That's a pretty simple set of tests, but we also only allow for create/update/delete, so that is all we need to test.
To run that testing in Travis we have to set up Minikube via this [script](../test-scripts/cluster-up-minikube.sh)
before actually running the integration tests. It should be noted that we need minikube 0.22.3 or above due to a bug in
0.22.2 and below which caused issues with kubectl interactions.

### Running Integration Tests Locally

To run integration tests locally you need to run `minikube start` to get a minikube context, and then run
`make integration-test`. You will need to have installed the necessary requirements by running
`pip3 install requirements.txt`. Testing is done via a Python module using nosetests.