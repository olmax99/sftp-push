# Build Stage
FROM ubuntu-sftp-push:1.13 AS build-stage

LABEL app="build-sftppush"
LABEL REPO="https://github.com/olmax99/sftppush"

ENV PROJPATH=/go/src/github.com/olmax99/sftppush

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

ADD . /go/src/github.com/olmax99/sftppush
WORKDIR /go/src/github.com/olmax99/sftppush

RUN make build-alpine

# Final Stage
FROM ubuntu:16.04

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/olmax99/sftppush"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/sftppush/bin

WORKDIR /opt/sftppush/bin

COPY --from=build-stage /go/src/github.com/olmax99/sftppush/bin/sftppush /opt/sftppush/bin/
RUN chmod +x /opt/sftppush/bin/sftppush

# Create appuser
RUN adduser -D -g '' sftppush
USER sftppush

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["/opt/sftppush/bin/sftppush"]
