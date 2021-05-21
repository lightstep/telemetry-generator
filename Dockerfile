FROM golang:1.15-stretch as builder

RUN mkdir /build
WORKDIR /build

RUN GO111MODULE=on go get github.com/open-telemetry/opentelemetry-collector-builder

ADD . .

RUN /go/bin/opentelemetry-collector-builder --config /build/builder-config.yml

FROM debian:stretch-slim

RUN mkdir -p /etc/otel
WORKDIR /otel

COPY --from=builder /tmp/ls-partner-col-distribution/lightstep-partner-collector .

ENTRYPOINT [ "./lightstep-partner-collector" ]
CMD [ "--config", "/etc/otel/config.yaml" ]