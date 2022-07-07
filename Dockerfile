FROM golang:1.17 AS builder
WORKDIR /go/src/open-cluster-management.io/addon-tutorial
COPY . .
ENV GO_PACKAGE open-cluster-management.io/addon-tutorial

ENV GOFLAGS ""
RUN go env
RUN go build -a -o busyboxaddon examples/busyboxaddon/manager/main.go


FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
COPY --from=builder /go/src/open-cluster-management.io/addon-tutorial/busyboxaddon /

RUN microdnf update && microdnf clean all
