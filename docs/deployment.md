# Deployment

We don't deploy a docker image with Lostromos, but instead only release the binary to github so that it can be built in
to the image you need for your particular application. This is due to the link between kubectl and Lostromos, and it's
possible the kubectl we intend to use isn't valid for everyones use case. We do build a docker image as part of testing
that can be thought of as an [example](../test/docker/Dockerfile) however.

## GOOS Environment Variable

In order for `lostromos` to work in a scratch/alpine container (and therefore be as small as possible), we need to build
the binary with the normal `make build` command, but having set GOOS=linux.

## Kubectl Version

We test with [kubectl] version 1.7.x for our installs due to an issue with 1.8 that causes error messages
(false failures).

[kubectl]: https://kubernetes.io/docs/user-guide/kubectl-overview/