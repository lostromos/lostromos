# Development

In order to do development locally, you should have the version of Golang supported by this project, and follow the
given instructions to setup your environment.

## Install Deps

By running `make install-deps` you should be able to get a setup locally to do anything related to this project.
Anything go related will install to your GOPATH, and anything else will have a docker image used to perform tasks.
Docker is a requirement for working with this project as well as Golang. You can also set up your vendor directory by
running `make vendor`.

## Git Hooks

For an automated way to ensure linting is run before a commit is pushed, you can run `./git-hooks/install` to set up a
pre-commit hook that will run linting. This isn't a requirement, just a nice to have to find issues before they get to
Travis.

## Testing Locally

To run a functional test on your own to validate that the lostromos app is working as expected you can go through the
following steps.

1. Setup kubectl against a cluster (minikube works just fine)
2. `kubectl apply -f test-data/crd.yml`
3. `kubectl apply -f test-data/cr_things.yml`
4. `go run main.go start --config test-data/config.yaml`
    - See that it prints out that `thing1` and `thing2` were added
5. In another shell `kubectl apply -f test-data/cr_nemo.yml`
    - See that it prints out that `nemo` was added
6. `kubectl edit character nemo` and change a field
    - See that it prints out that `nemo` was changed
7. `kubectl delete -f test-data/cr_nemo.yml`
    - See that it prints out that `nemo` was deleted
8. `kubectl delete -f test-data/cr_things.yml`
    - See that it prints out that `thing1` and `thing2` were deleted
9. That's it. You can stop the process and `kubectl delete -f test-data/crd.yml` to cleanup the rest of the test data.

This also happens to be what we do in the [Integration Tests](./../test-scripts/integration-tests.sh) to ensure we are
working as expected after every build. Checkout our [Testing](./testing.md) documentation for more information on that
script as well as unit tests.