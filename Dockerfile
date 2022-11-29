FROM golang:1.19.1 as builder

RUN mkdir /build
WORKDIR /build

# Default tag is demo
ARG TAG=local-dev

RUN GO111MODULE=on go install go.opentelemetry.io/collector/cmd/builder@v0.60.0 

ADD . .
# Substitute TAG into config file as version
RUN sed -i -e "s/\${TAG}/$TAG/g" /build/config/builder-config.yml 

RUN builder --config /build/config/builder-config.yml 

FROM debian:stretch-slim

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates

RUN update-ca-certificates

RUN mkdir -p /etc/otel
WORKDIR /otel

COPY --from=builder /build/build/telemetry-generator .
COPY --from=builder /build/config/collector-config.yml /etc/otel/config.yaml
COPY --from=builder /build/examples/* /otel/examples/

ENV TOPO_FILE=/otel/examples/hipster_shop.yaml
ENV OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=ingest.lightstep.com:443
ENV OTEL_EXPORTER_OTLP_TRACES_ENDPOINT_INTERNAL=ingest.lightstep.com:443
ENV OTEL_INSECURE=false

ENTRYPOINT [ "./telemetry-generator" ]
CMD [ "--config", "/etc/otel/config.yaml" ]
