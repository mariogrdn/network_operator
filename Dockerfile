############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git
COPY operator.go ./operator.go
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
ENTRYPOINT [ "go", "run", "./operator.go" ]
