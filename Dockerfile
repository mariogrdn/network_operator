############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git wireless-tools
WORKDIR $GOPATH/network_operator/operator/
COPY . .
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
RUN go build -o /go/bin/operator
############################
# STEP 2 build a small image
############################
FROM alpine 
# Copy our static executable.
COPY --from=builder /go/bin/operator /go/bin/operator
RUN apk update && apk add --no-cache wireless-tools
# Run the server binary.
ENTRYPOINT ["/go/bin/operator"]
