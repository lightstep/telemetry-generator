package generatorreceiver

import (
	"go.opentelemetry.io/collector/config/confighttp"
)

// Config defines configuration for OTLP receiver.
type Config struct {
	// Path of generator config file. Path is relative to current directory.
	Path string `mapstructure:"path"`
	// Inline string containing the topo file
	InlineFile string `mapstructure:"inline"`
	// ApiIngress holds config settings for HTTP server listening for requests.
	ApiIngress confighttp.HTTPServerSettings `mapstructure:"api"`
}
