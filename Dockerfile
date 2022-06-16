FROM golang:1.18

COPY . /go/src/github.com/cyverse-de/jex-adapter
WORKDIR /go/src/github.com/cyverse-de/jex-adapter
ENV CGO_ENABLED=0
RUN go build --buildvcs=false .


FROM debian:stable-slim

WORKDIR /app

COPY --from=0 /go/src/github.com/cyverse-de/jex-adapter/jex-adapter /bin/jex-adapter

ENTRYPOINT ["jex-adapter"]

EXPOSE 60000
