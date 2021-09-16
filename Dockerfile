FROM golang:1.17-stretch as builder

RUN mkdir /build
WORKDIR /build

RUN GO111MODULE=on go get github.com/open-telemetry/opentelemetry-collector-builder@v0.30.0

ADD . .

RUN /go/bin/opentelemetry-collector-builder --config /build/builder-config.yml --otelcol-version 0.30.0

FROM debian:stretch-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates

RUN update-ca-certificates

RUN mkdir -p /etc/otel
WORKDIR /otel

COPY --from=builder /tmp/ls-partner-col-distribution/lightstep-partner-collector .
COPY --from=builder /build/generatorreceiver/topos/hipster_shop.json /etc/otel/hipster_shop.json
COPY --from=builder /build/config/collector-config.yml /etc/otel/config.yaml

ENV TOPO_FILE=/etc/otel/hipster_shop.json
ENV OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=ingest.lightstep.com:443
ENV OTEL_INSECURE=false

ENTRYPOINT [ "./lightstep-partner-collector" ]
CMD [ "--config", "/etc/otel/config.yaml" ]