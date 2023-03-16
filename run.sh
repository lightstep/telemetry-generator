#!/bin/bash

export LS_ACCESS_TOKEN=dontcare1
export LS_ACCESS_TOKEN_INTERNAL=dontcare2
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT=127.0.0.1:4318
export OTEL_EXPORTER_OTLP_TRACES_ENDPOINT_INTERNAL=127.0.0.1:4318
export OTEL_INSECURE=true
export TOPO_FILE=/Users/josh.macdonald/src/opentelemetry/telemetry-generator/examples/hipster_shop.yaml

make run-local
