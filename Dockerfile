FROM golang:1.15-stretch as builder

RUN mkdir /build
WORKDIR /build

RUN GO111MODULE=on go get github.com/open-telemetry/opentelemetry-collector-builder

ADD . .

RUN /go/bin/opentelemetry-collector-builder --config /build/builder-config.yml --otelcol-version 0.30.0

FROM debian:stretch-slim

RUN mkdir -p /etc/otel
WORKDIR /otel

COPY --from=builder /tmp/ls-partner-col-distribution/lightstep-partner-collector .
COPY --from=builder /build/generatorreceiver/topos/hipster_shop.json /etc/otel/hipster_shop.json
COPY --from=builder /build/config/collector-config.yml /etc/otel/config.yaml

ENV TOPO_FILE=/etc/otel/hipster_shop.json

ENTRYPOINT [ "./lightstep-partner-collector" ]
CMD [ "--config", "/etc/otel/config.yaml" ]