FROM golang:alpine AS build-env

# Install any compile-time system dependencies.
RUN apk add --no-cache git curl
RUN go get -u github.com/golang/dep/...
ENV KUBECTL_VERSION v1.7.9
RUN curl -L -o /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl
RUN chmod +x /usr/bin/kubectl

# Copy lostromos into the build environment.
WORKDIR /go/src/github.com/wpengine/lostromos
COPY .  /go/src/github.com/wpengine/lostromos

# Install any compile-time golang dependencies.
RUN dep ensure
RUN CGO_ENABLED=0 go install github.com/wpengine/lostromos

FROM alpine:latest

# Add a lostromos user
RUN adduser -D lostromos
USER lostromos

# Add our compiled binary and kubectl
COPY --from=build-env /go/bin/lostromos /lostromos
COPY --from=build-env /usr/bin/kubectl /usr/bin/kubectl

ENTRYPOINT ["/lostromos"]
