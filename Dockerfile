FROM golang:1.19.1 as builder

RUN mkdir /build
WORKDIR /build

RUN GO111MODULE=on go install github.com/open-telemetry/opentelemetry-collector-builder@v0.35.0

ADD . .

RUN /go/bin/opentelemetry-collector-builder --config /build/config/builder-config.yml

FROM debian:stretch-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates

RUN update-ca-certificates

RUN mkdir -p /etc/otel
WORKDIR /otel

COPY --from=builder /build/build/telemetry-generator .
COPY --from=builder /build/examples/* /etc/otel/
COPY --from=builder /build/config/collector-config.yml /etc/otel/config.yaml

ENV TOPO_FILE=/etc/otel/hipster_shop.yaml
ENV OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=ingest.lightstep.com:443
ENV OTEL_INSECURE=false

ENTRYPOINT [ "./telemetry-generator" ]
CMD [ "--config", "/etc/otel/config.yaml" ]
