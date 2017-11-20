#!/bin/bash

NSENTER_IMAGE=jpetazzo/nsenter

check_or_build_nsenter() {
    which nsenter >/dev/null && return 0
    echo "INFO: Building 'nsenter' ..."
    docker pull $NSENTER_IMAGE
    docker run --rm -v `pwd`:/target $NSENTER_IMAGE
    if [ ! -f ./nsenter ]; then
        echo "ERROR: nsenter pull failed, log:"
        return 1
    fi
    echo "INFO: nsenter build OK, installing ..."
    install_bin ./nsenter
}

install_bin() {
    local exe=${1:?}
    test -n "${TRAVIS}" && sudo install -v ${exe} /usr/local/bin || install ${exe} ${GOPATH:?}/bin
}

check_or_build_nsenter