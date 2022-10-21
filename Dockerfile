FROM golang:1.17 AS builder
WORKDIR /go/src/open-cluster-management.io/addon-tutorial
COPY . .
ENV GO_PACKAGE open-cluster-management.io/addon-tutorial

ENV GOFLAGS ""
RUN go env
RUN go build -a -o busybox-addon examples/busyboxaddon/manager/main.go
RUN	go build -a -o leaseprober-addon examples/leaseproberaddon/manager/main.go
RUN	go build -a -o leaseprober-agent examples/leaseproberaddon/agent/main.go
RUN	go build -a -o workprober-addon examples/workproberaddon/manager/main.go
RUN go build -a -o large-addon examples/large-addon/agent/main.go


FROM registry.access.redhat.com/ubi8/ubi-minimal:latest
COPY --from=builder /go/src/open-cluster-management.io/addon-tutorial/busybox-addon /
COPY --from=builder /go/src/open-cluster-management.io/addon-tutorial/leaseprober-addon /
COPY --from=builder /go/src/open-cluster-management.io/addon-tutorial/leaseprober-agent /
COPY --from=builder /go/src/open-cluster-management.io/addon-tutorial/workprober-addon /
COPY --from=builder /go/src/open-cluster-management.io/addon-tutorial/large-addon /


RUN microdnf update && microdnf clean all
