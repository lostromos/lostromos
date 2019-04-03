FROM golang:alpine AS build-env

# Install any compile-time system dependencies.
RUN apk add --no-cache git curl make
RUN go get -u github.com/golang/dep/...
ENV KUBECTL_VERSION v1.9.2
RUN curl -L -o /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl
RUN chmod +x /usr/bin/kubectl

# Copy lostromos into the build environment.
WORKDIR /go/src/github.com/lostromos/lostromos
COPY .  /go/src/github.com/lostromos/lostromos

# Install any compile-time golang dependencies.
RUN dep ensure
RUN make out/lostromos-linux-amd64

FROM alpine:latest

# Add a lostromos user
RUN adduser -D lostromos
USER lostromos

# Add our compiled binary and kubectl
COPY --from=build-env /go/src/github.com/lostromos/lostromos/out/lostromos-linux-amd64 /lostromos
COPY --from=build-env /usr/bin/kubectl /usr/bin/kubectl

ENTRYPOINT ["/lostromos"]
