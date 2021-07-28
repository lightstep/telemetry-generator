package generatorreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
	"time"
)

const (
	typeStr         = "generator"
	DefaultTopoFile = "topo.json"
)

// NewFactory creates a factory for AWS receiver.
func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithTraces(createTracesReceiver),
		receiverhelper.WithMetrics(createMetricsReceiver))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewID(typeStr)),
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