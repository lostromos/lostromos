# Development

## Dependencies

1. Golang 1.9.x
2. Minikube >0.22.3
3. Docker
4. Python3

## Make Dependency Targets

Some make targets explicitly leave off dependencies as part of the build since they are slow. For instance targets like
`install-go-deps` only needs to be run once when setting up the project. `install-go-deps` is technically a dependency
of the `build` target, but you probably don't want to have to run it before every `make build` since you will already
have all the dependencies and running the target will just add several seconds to your build for no gain. Instead run
this when you first start on the project/a new PR

```bash
make install-go-deps
make vendor
make install-python-deps
```

## Git Hooks

For an automated way to ensure linting is run before a commit is pushed, you can run `./git-hooks/install` to set up a
pre-commit hook that will run linting. This isn't a requirement, just a nice to have to find issues before they get to
Travis.

## Testing Locally

To run a functional test on your own to validate that the lostromos app is working as expected you can go through the
following steps.

1. Setup kubectl against a cluster (minikube works just fine)
2. `kubectl apply -f test/data/crd.yml`
3. `kubectl apply -f test/data/cr_things.yml`
4. `go run main.go start --config test/data/config.yaml`
  - See that it prints out that `thing1` and `thing2` were added
5. In another shell `kubectl apply -f test/data/cr_nemo.yml`
  - See that it prints out that `nemo` was added
6. `kubectl edit character nemo` and change a field
  - See that it prints out that `nemo` was changed
7. `kubectl delete -f test/data/cr_nemo.yml`
  - See that it prints out that `nemo` was deleted
8. `kubectl delete -f test/data/cr_things.yml`
  - See that it prints out that `thing1` and `thing2` were deleted
9. You can stop the process and `kubectl delete -f test/data/crd.yml` to cleanup the rest of the test data.

This also happens to be what we do in the [Integration Tests](./../test/scripts/integration-tests.sh) to ensure we are
working as expected after every build. Checkout our [Testing](./testing.md) documentation for more information on that
script as well as unit tests.

## Using the Test Docker Image

If you are trying to run testing on the docker image, the easy way to link your minikube context with the image is to
run `eval $(minikube docker-env)` and build the test image found [here](../test/docker/Dockerfile). You might want to
upgrade to a version of minikube 0.23.0 or above like we do in our integration tests. Once you have an image and a
Minikube context you should run

```bash
kubectl create -f test/data/deploy.yaml
kubectl expose pod lostromos --type=LoadBalancer
curl `minikube service lostromos --url`/status
```

This will create a lostromos service in your minikube environment. Alternatively you can run `make docker-build-test`.