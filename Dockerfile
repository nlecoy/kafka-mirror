FROM golang:1.17-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY main.go ./

RUN go build -ldflags "-extldflags \"-static\"" -o release/kafka-mirror github.com/nlecoy/kafka-mirror

FROM alpine:3.11

COPY --from=build /app/release/kafka-mirror /bin/

ENTRYPOINT [ "/bin/kafka-mirror" ]
