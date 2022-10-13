package generatorreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"time"
)

const (
	typeStr         = "generator"
	DefaultTopoFile = "topo.json"
	// The stability level of the exporter.
	stability = component.StabilityLevelStable
)

// NewFactory creates a factory for the receiver.
func NewFactory() component.ReceiverFactory {
	return component.NewReceiverFactory(
		typeStr,
		createDefaultConfig,
		component.WithTracesReceiver(createTracesReceiver, stability),
		component.WithMetricsReceiver(createMetricsReceiver, stability))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewComponentID(typeStr)),
		Path:             DefaultTopoFile,
	}
}

func createMetricsReceiver(
	ctx context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	consumer consumer.Metrics) (component.MetricsReceiver, error) {
	rcfg := cfg.(*Config)
	return newMetricReceiver(rcfg, consumer, params.Logger, time.Now().Unix())
}

func createTracesReceiver(
	ctx context.Context,
	params component.ReceiverCreateSettings,
	cfg config.Receiver,
	consumer consumer.Traces) (component.TracesReceiver, error) {
	rcfg := cfg.(*Config)
	return newTraceReceiver(rcfg, consumer, params.Logger, time.Now().Unix())
}
