![Lostrómos logo](images/logo.png)

# <a name="contributing"></a>Contributing to Lostrómos
## <a id="development"></a>Development
### Make Dependency Targets

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

### Git Hooks

For an automated way to ensure linting is run before a commit is pushed, you can run `./git-hooks/install` to set up a
pre-commit hook that will run linting. This isn't a requirement, just a nice to have to find issues before they get to
Travis.

### Using the Test Docker Image

If you are trying to run testing on the docker image, the easy way to link your minikube context with the image is to
run `eval $(minikube docker-env)` and build the test image found [here](../test/docker/Dockerfile). Once you have an
image and a Minikube context you should run

```bash
kubectl create -f test/data/deploy.yaml
kubectl expose pod lostromos --type=LoadBalancer
curl `minikube service lostromos --url`/status
```

This will create a lostromos service in your minikube environment. Alternatively you can run `make docker-build-test`.

## <a name="testing"></a>Testing

### Unit Testing

This codebase is expected to be highly tested via unit testing. New functionality should have a unit test if at all
possible, and if not possible that should be explained in the original PR. We may ask you to rewrite the code so that it
is testable, and PRs without tests are more likely to be rejected. Before submitting a PR, run `make test` to ensure no
unit tests have started to fail.

### Code Coverage

Code Coverage is expected to be at least 80%, with a stretch goal of 100%. We track our code coverage via
[codecov.io](https://codecov.io/gh/wpengine/lostromos). To generate a coverage report locally, you can run
`make coverage`.

### Integration Testing

To ensure that Lostromos is working as expected we do integration testing through
[minikube](https://github.com/kubernetes/minikube) to ensure we can watch resources related to our custom resource
definition. We test that

1. We can see resources that already exist.
2. Verify an update has occurred (no processing, just can see it happens).
3. Alert when resources are being deleted.

These tests cover the create/update/delete use cases that lostromos allows. To run that testing in Travis we have to set
up Minikube via this [script](../test/scripts/install_minikube.sh) before actually running the integration tests.

#### Testing Docker Image

Integration testing is run against a docker image with kubectl 1.7.x and lostromos installed to ensure that we have a
working lostromos binary. We push the image as a service in minikube to be able to ping the /metrics and /status urls
for use in the tests.

#### Running Integration Tests Locally

To run integration tests locally you need to run `minikube start` to get a minikube context, and then run
`make integration-test`. This is assuming you have installed the necessary python requirements via
`make install-python-deps`.

## <a name="ci"></a>Continuous Integration

### Working with [Travis CI](https://travis-ci.org/)

We use Travis as our continuous integration environment. In order to have a PR available for merging, it must pass the
tests run as part of the CI pipeline. The same thing applies for a new build, which must pass testing against the master
branch before deploying a new image.

### How to run Travis

Travis uses a yml system of running tasks. Check out the [.travis.yml](../.travis.yml) file to see more detailed
information about what operations we run. The travis job is hosted [here](https://travis-ci.org/wpengine/lostromos).

### Slow Builds

We use the open source version of TravisCI, which means we are beholden to whatever state of
the system Travis is in at that point. You can go [here](https://www.traviscistatus.com/) to check out the status of
Travis as well as some metrics on how many other builds are currently going. In the case where your build is going
slowly, or it's sitting in the queue for an extended period, the metrics at the bottom of the page might give you an
idea as to why. 

### Spell Out Build Steps

When adding to the travis.yml file, we want to spell out the steps we are taking to run the build. What that means
practically is that we don't create make targets specifically for Travis. We use the make targets that exist for local
usage, things like `make lint`, `make test`, `make build`, etc... because those are targets created for local use that
should be done the exact same way in Travis. We could go the route of having a target `travis-integration-tests` that
looks similar to

```makefile
travis-integration-tests: lint build integration-tests
```

but that wouldn't spell out what we are trying to do when looking at build logs, and we also wouldn't get step by step
timing information since it is all one step. For those reasons we've decided to spell things out instead of abstracting
the details away.

## [Contribution Guidelines](../CONTRIBUTING.md)

## 
