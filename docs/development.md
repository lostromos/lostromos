# Development

In order to do development locally, you should have the version of Golang supported by this project, and follow the
given instructions to setup your environment.

## Install Deps

By running `make install-deps` you should be able to get a setup locally to do anything related to this project.
Anything go related will install to your GOPATH, and anything else will have a docker image used to perform tasks.
Docker is a requirement for working with this project as well as Golang.

## Git Hooks

For an automated way to ensure linting is run before a commit is pushed, you can run `./git-hooks/install` to set up a
pre-commit hook that will run linting. This isn't a requirement, just a nice to have to find issues before they get to
Travis.