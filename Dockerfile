ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARC=amd64
ARG GOPATH=/opt
ARG USERNAME=go
#################################
# STEP 1 build the static binary
################################
FROM golang:latest as builder
ARG CGO_ENABLED
ARG GOOS
ARG GOARC
ARG GOPATH
ARG USERNAME
RUN apt-get update && apt-get install ca-certificates
RUN useradd -m ${USERNAME}
WORKDIR /opt
COPY . .
RUN CGO_ENABLED=${CGO_ENABLED} GOOS=${GOOS} GOARC=${GOARC} GOPATH=${GOPATH} go build -a -o /go/bin/canihazconnection

#########################
# STEP 2 build the image
#########################
FROM scratch
ARG USERNAME
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/canihazconnection /canihazconnection
USER ${USERNAME}
CMD ["/canihazconnection"]