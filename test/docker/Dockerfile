FROM alpine
WORKDIR /

# Install curl so that we can download kubectl
RUN apk add --update curl && \
    rm -rf /var/cache/apk/*

# Add kubectl to the image
ENV KUBECTL_VERSION v1.7.9
RUN curl -L -o /usr/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl
RUN chmod +x /usr/bin/kubectl

# Setup the Lostromos binary.
EXPOSE 8080
ADD ./lostromos /
COPY ./test/data /test/data
ENTRYPOINT ["/lostromos"]