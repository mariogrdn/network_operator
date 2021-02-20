############################
# STEP 1 build executable binary
############################
FROM partlab/ubuntu-golang AS builder
# Install git.
# Git is required for fetching the dependencies.
RUN apt-get update && apt-get install -y git
RUN apt-get install -y wireless-tools
COPY operator.go ./operator.go
ENV GO111MODULE=off
# Fetch dependencies.
# Using go get.
RUN go get -d -v
# Build the binary.
ENTRYPOINT [ "go", "run", "./operator.go" ]
