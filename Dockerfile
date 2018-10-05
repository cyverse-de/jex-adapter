FROM golang:1.11-alpine

RUN apk add --no-cache git
RUN go get -u github.com/jstemmer/go-junit-report

COPY . /go/src/github.com/cyverse-de/jex-adapter
ENV CGO_ENABLED=0
RUN go install github.com/cyverse-de/jex-adapter

ENTRYPOINT ["jex-adapter"]
CMD ["--help"]
EXPOSE 60000

ARG git_commit=unknown
ARG version="2.9.0"
ARG descriptive_version=unknown

LABEL org.cyverse.git-ref="$git_commit"
LABEL org.cyverse.version="$version"
LABEL org.cyverse.descriptive-version="$descriptive_version"
LABEL org.label-schema.vcs-ref="$git_commit"
LABEL org.label-schema.vcs-url="https://github.com/cyverse-de/jex-adapter"
LABEL org.label-schema.version="$descriptive_version"
