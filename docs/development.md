# Development

This page is meant for more in depth information about doing development on Lostrómos. If you came here first, you might
want to look at the [Quick Start](https://github.com/wpengine/lostromos#quick-start) from the README before going
further.

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

## Using the Test Docker Image

If you are trying to run testing on the docker image, the easy way to link your minikube context with the image is to
run `eval $(minikube docker-env)` and build the test image found [here](../test/docker/Dockerfile). Once you have an
image and a Minikube context you should run

```bash
kubectl create -f test/data/deploy.yaml
kubectl expose pod lostromos --type=LoadBalancer
curl `minikube service lostromos --url`/status
```

This will create a lostromos service in your minikube environment. Alternatively you can run `make docker-build-test`.