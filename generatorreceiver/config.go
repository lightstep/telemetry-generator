package generatorreceiver

import "go.opentelemetry.io/collector/config"

// Config defines configuration for OTLP receiver.
type Config struct {
	config.ReceiverSettings `mapstructure:",squash"` // squash ensures fields are correctly decoded in embedded struct
	// Path of generator config file. Path is relative to current directory.
	Path string `mapstructure:"path"`
}
