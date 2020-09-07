####
# BUILD STAGE
####

FROM golang:1.13.15-buster AS build-stage

# ------ 1. label and Go-Path--------
LABEL app="sftppush"
LABEL REPO="https://github.com/olmax99/sftppush"

ENV PROJPATH=/go/src/github.com/olmax99/sftppush

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:$GOROOT/bin:$GOPATH/bin

# ------ 2. Get Go Project-----------
ADD . /go/src/github.com/olmax99/sftppush
WORKDIR /go/src/github.com/olmax99/sftppush

# ------ 3. Create Go Binary---------
RUN make build-alpine

####
# FINAL STAGE
#### 

FROM golang:1.13.15-buster

RUN apt update && apt install -y \
    dumb-init \
    ca-certificates \
    openssl

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
RUN groupadd -r sftppush && useradd -r -g sftppush sftppush
USER sftppush

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

# details at https://github.com/Yelp/dumb-init
# CMD ["bash", "-c", "do-some-pre-start-thing && exec my-server"]
CMD ["/opt/sftppush/bin/sftppush"]
