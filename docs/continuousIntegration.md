# Continuous Integration

## Working with Travis CI

We use Travis as our continuous integration environment. In order to have a PR available for merging, it must pass the
tests run as part of the CI pipeline. The same thing applies for a new build, which must pass testing against the master
branch before deploying a new image.

## How to run Travis

Travis uses a yml system of running tasks. Check out the [.travis.yml](../.travis.yml) file to see more detailed
information about what operations we run. The travis job is hosted [here](https://travis-ci.org/wpengine/lostromos).

### Slow Builds

We use the open source version of [TravisCI](https://travis-ci.org), which means we are beholden to whatever state of
the system Travis is in at that point. You can go [here](https://www.traviscistatus.com/) to check out the status of
Travis as well as some metrics on how many other builds are currently going. In the case where your build is going
slowly, or it's sitting in the queue for an extended period, the metrics at the bottom of the page might give you an
idea as to why. It doesn't seem there is anything we can do about this situation, so this is more about information then
proposing a solution.

## Spell Out Build Steps

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