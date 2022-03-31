FROM golang:latest AS builder
WORKDIR /

# Build the binary
ENV SRC_DIR=/src
ENV GO111MODULE=on
RUN mkdir -p ${SRC_DIR}
COPY go.* $SRC_DIR/
COPY *.go ${SRC_DIR}/
WORKDIR $SRC_DIR
RUN go mod download
RUN go build main.go
RUN mv main controller

# Setup the site data
FROM gcr.io/distroless/base:latest
COPY --from=builder /src/controller /

# Useful environment variables:
#
# * GITHUB_TOKEN: Sets GitHub API token
# * CONFIG_PATH: Sets configuration path (defaults to "/app/config/config.yaml")
# * PORT: Sets HTTP listening port (defaults to 8080)
# * PERSIST_BACKEND: Set the cache persistence backend
# * PERSIST_PATH: Set the cache persistence path
#
# For other environment variables, see:
# https://github.com/google/triage-party/blob/master/docs/deploy.md
CMD ["/controller"]
