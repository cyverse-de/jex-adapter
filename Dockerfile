FROM golang:1.7-alpine

RUN apk add --update git perl \
    && rm -rf /var/cache/apk

RUN go get github.com/jstemmer/go-junit-report

ENV CONSUL_TEMPLATE_VERSION=0.16.0
ENV CONSUL_TEMPLATE_SHA256SUM=064b0b492bb7ca3663811d297436a4bbf3226de706d2b76adade7021cd22e156
ENV CONSUL_CONNECT=localhost:8500
ENV CONF_TEMPLATE=/go/src/github.com/cyverse-de/jex-adapter/jobservices.yml.tmpl
ENV CONF_FILENAME=jobservices.yml
ENV PROGRAM=jex-adapter

ADD https://releases.hashicorp.com/consul-template/${CONSUL_TEMPLATE_VERSION}/consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip .

RUN echo "${CONSUL_TEMPLATE_SHA256SUM}  consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip" | sha256sum -c - \
    && unzip consul-template_${CONSUL_TEMPLATE_VERSION}_linux_amd64.zip \
    && mkdir -p /usr/local/bin \
    && mv consul-template /usr/local/bin/consul-template

COPY run-service /usr/local/bin/run-service

COPY . /go/src/github.com/cyverse-de/jex-adapter
RUN go install github.com/cyverse-de/jex-adapter

EXPOSE 60000
ENTRYPOINT ["run-service"]
CMD ["--help"]

ARG git_commit=unknown
ARG version="2.9.0"

LABEL org.cyverse.git-ref="$git_commit"
LABEL org.cyverse.version="$version"
