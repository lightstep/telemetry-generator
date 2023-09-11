package generatorreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"time"
)

const (
	typeStr         = "generator"
	DefaultTopoFile = "topo.json"
	// The stability level of the exporter.
	stability = component.StabilityLevelStable
)

// NewFactory creates a factory for the receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithTraces(createTracesReceiver, stability),
		receiver.WithMetrics(createMetricsReceiver, stability))
}

func createDefaultConfig() component.Config {
	return &Config{
		Path: DefaultTopoFile,
	}
}

func createMetricsReceiver(
	ctx context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Metrics) (receiver.Metrics, error) {
	rcfg := cfg.(*Config)
	return newMetricReceiver(rcfg, consumer, params.Logger, time.Now().Unix())
}

func createTracesReceiver(
	ctx context.Context,
	params receiver.CreateSettings,
	cfg component.Config,
	consumer consumer.Traces) (receiver.Traces, error) {
	rcfg := cfg.(*Config)
	return newTraceReceiver(rcfg, consumer, params.Logger, time.Now().Unix())
}
