module github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver

go 1.16

require go.opentelemetry.io/collector v0.35.0

require (
	github.com/robfig/cron/v3 v3.0.1
	github.com/stretchr/testify v1.8.0 // indirect
	go.opentelemetry.io/collector/model v0.35.0
	go.opentelemetry.io/otel v1.0.0-RC3
	go.uber.org/zap v1.19.1
	gopkg.in/yaml.v3 v3.0.1
)
