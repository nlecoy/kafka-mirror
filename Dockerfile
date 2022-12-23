FROM reg.b47ch.com/app-go-builder:1.16-alpine AS build

COPY vendor /go/src/github.com/nlecoy/kafka-mirror/vendor

RUN go install github.com/nlecoy/kafka-mirror/vendor/...

COPY *.go /go/src/github.com/nlecoy/kafka-mirror/

RUN go install github.com/nlecoy/kafka-mirror/...

FROM reg.b47ch.com/app-go-base:alpine

COPY --from=build /go/bin/kafka-mirror /go/bin/kafka-mirror

ENTRYPOINT ["/go/bin/kafka-mirror"]
