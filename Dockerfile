FROM golang:alpine3.13 as builder

RUN apk add --no-cache git

RUN mkdir /build
RUN go get github.com/open-telemetry/opentelemetry-collector-builder@v0.8.0

WORKDIR /build
ADD . .

WORKDIR /build/backstageprocessor
RUN go mod download

WORKDIR /build/webhookprocessor
RUN go mod download

WORKDIR /build

RUN /go/bin/opentelemetry-collector-builder --config /build/builder-config.yml

FROM alpine:3.13
COPY --from=builder /build/config .
COPY --from=builder /tmp/ls-partner-col-distribution/lightstep-partner-collector .

ENTRYPOINT [ "./lightstep-partner-collector" ]
CMD [ "--config", "./config/collector-config.yml" ]