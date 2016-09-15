FROM golang:1.7-alpine

ARG git_commit=unknown
ARG version="2.8.1"

LABEL org.cyverse.git-ref="$git_commit"
LABEL org.cyverse.version="$version"

COPY . /go/src/github.com/cyverse-de/jex-adapter
RUN go install github.com/cyverse-de/jex-adapter

EXPOSE 60000
ENTRYPOINT ["jex-adapter"]
CMD ["--help"]
