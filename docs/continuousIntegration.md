# Continuous Integration

## Working with Travis CI

We use Travis as our continuous integration environment. In order to have a PR available for merging, it must pass the
tests run as part of the CI pipeline. The same thing applies for a new build, which must pass testing against the master
branch before deploying a new image.

## How to run Travis

Travis uses a yml system of running tasks. Check out the [.travis.yml](../.travis.yml) file to see more detailed
information about what operations we run. The travis job is hosted [here](https://travis-ci.org/wpengine/lostromos).
