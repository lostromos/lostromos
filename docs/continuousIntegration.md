# Continuous Integration

## Working with Travis CI

We use Travis as our continuous integration environment. In order to have a PR available for merging, it must pass the
tests run as part of the CI pipeline. The same thing applies for a new build, which must pass testing against the master
branch before deploying a new image.

## How to run Travis

Travis uses a yml system of running tasks. Check out the [.travis.yml](../.travis.yml) file to see more detailed
information about what operations we run. The travis job is hosted [here](https://travis-ci.org/wpengine/lostromos).

## Multiple Builds

We use a build matrix consisting of Go version 1.9 and the latest master for unit testing and integration testing. That
is 4 possible jobs running, where we don't fail if the latest master for Go causes an issue. We merely want to see how
our framework works with it to alert to any future problems, not fail. Both unit testing and integration testing perform
builds, and ensure linting, but only the unit testing does code coverage currently.

## Spell Out Build Steps

When adding to the travis.yml file, we want to spell out the steps we are taking to run the build. What that means
practically is that we don't create make targets specifically for Travis. For example in the `script` portion of the
travis.yml file we have an if/else clause for integration vs unit testing. We use the make targets that exist for local
usage, things like `make lint`, `make test`, `make build`, etc... because those are targets created for local use that
should be done the exact same way in Travis. We could go the route of having a target `travis-integration-tests` that
looks similar to

```makefile
travis-integration-tests: lint build integration-tests
```

but that wouldn't spell out what we are trying to do when looking at build logs, and we also wouldn't get step by step
timing information since it is all one step. For those reasons we've decided to spell things out instead of abstracting
the details away.