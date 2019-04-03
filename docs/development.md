![Lostrómos logo](images/logo.png)

# <a name="contributing"></a>Contributing to Lostrómos

## <a id="development"></a>Development

### New Go and Python Dependencies

Go dependencies are managed via [dep](https://github.com/golang/dep).

* Run `dep ensure` to add your new Go dependencies

Python dependencies are managed via [pip-compile](https://github.com/jazzband/pip-tools)

* Add your Python dependencies to [requirements.in](../requirements.in)
* Run `pip-compile` to add your new Python dependencies

### Git Hooks

Our [CI](#ci) tool fails PRs that do not pass linting. For an automated way to
ensure linting is run before a commit is pushed, you can run
`./git-hooks/install` to set up a pre-commit hook that will run linting.

## <a name="testing"></a>Testing

### Unit Testing

This codebase is expected to be highly tested via unit testing. New
functionality should have a unit test if at allpossible, and if not possible
that should be explained in the original PR. We may ask you to rewrite the code
so that it is testable, and PRs without tests are more likely to be rejected.
Before submitting a PR, run `make test` to ensure no unit tests have started to
fail.

### Code Coverage

Code Coverage is expected to be at least 80%, with a stretch goal of 100%. We
track our code coverage via [codecov.io](https://codecov.io/gh/lostromos/lostromos).
To generate a coverage report locally, you can run `make coverage`.

### Integration Testing

To ensure that Lostrómos is working as expected we do integration testing
through Minikube. We test that

1. We can see resources that already exist.
2. We can verify an update has occurred.
3. We are alerted when resources are being deleted.

These tests cover the create/update/delete use cases that Lostrómos allows. Steps
to run the integration testing in Travis is found here: [travis.yaml](../.travis.yml)

#### Testing Docker Image

Integration testing is run against a docker image with kubectl 1.7.x and
Lostrómos installed to ensure that we have a working Lostrómos binary. We push
the image as a service in Minikube to be able to ping the /metrics and /status
urls for use in the tests.

#### Running Integration Tests Locally

To run integration tests locally you need to run `minikube start` to get a
Minikube context, and then run `make integration-test`. This is assuming you
have installed the necessary python requirements via `make install-python-deps`
and install nosetests: `sudo pip install nose`. Ignore error messages like
`TILLER...connect: connection refused` when running integration tests locally.

## <a name="ci"></a>Continuous Integration

### Working with [Travis CI](https://travis-ci.org/)

We use Travis as our continuous integration environment. In order to have a PR
available for merging, it must pass the tests run as part of the CI pipeline.
The same thing applies for a new build, which must pass testing against the
master branch before deploying a new image.

### How to run Travis

Travis uses a yml system of running tasks. Check out the [.travis.yml](../.travis.yml)
file to see more detailed information about what operations we run. The travis
job is hosted [here](https://travis-ci.org/lostromos/lostromos).

### Slow Builds

We use the open source version of TravisCI, which means we are beholden to
whatever state of the system Travis is in at that point. You can go [here](https://www.traviscistatus.com/)
to check out the status of Travis as well as some metrics on how many other
builds are currently going.

### Spell Out Build Steps

When adding to the travis.yml file, we want to spell out the steps we are taking
to run the build. What that means practically is that we don't create make
targets specifically for Travis. We use the make targets that exist for local
usage, things like `make lint`, `make test`, `make build`, etc... because those
are targets created for local use that should be done the exact same way in
Travis. We could go the route of having a target `travis-integration-tests` that
looks similar to

```makefile
travis-integration-tests: lint build integration-tests
```

but that wouldn't spell out what we are trying to do when looking at build logs,
and we also wouldn't get step by step timing information since it is all one
step. For those reasons we've decided to spell things out instead of abstracting
the details away.

## Contribution Guidelines

Ready to submit your contribution? See our [Contribution Guidelines](../CONTRIBUTING.md)
