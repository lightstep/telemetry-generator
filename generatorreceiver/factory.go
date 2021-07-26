package generatorreceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

const (
	typeStr = "generator"
	DefaultTopoFile = "topo.json"
)

// NewFactory creates a factory for AWS receiver.
func NewFactory() component.ReceiverFactory {
	return receiverhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		receiverhelper.WithTraces(createTracesReceiver))
}

func createDefaultConfig() config.Receiver {
	return &Config{
		ReceiverSettings: config.NewReceiverSettings(config.NewID(typeStr)),
		Path: DefaultTopoFile,
	}
}

func createTracesReceiver(
	ctx context.Context,
	params component.ReceiverCreateParams,
	cfg config.Receiver,
	consumer consumer.Traces) (component.TracesReceiver, error) {
	rcfg := cfg.(*Config)
	return newReceiver(rcfg, consumer, params.Logger)
}