package generatorreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/generator"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"time"
)

type generatorReceiver struct {
	logger     *zap.Logger
	traceConsumer   consumer.Traces
	metricConsumer   consumer.Metrics
	topoPath   string
	topoInline    string
	randomSeed int64
	metricGen  *generator.MetricGenerator
	tickers    []*time.Ticker
}

func (g generatorReceiver) loadTopoFile(topoInline string, path string) (*topology.File, error) {
	var topoFile topology.File

	// fetch from env var.
	if len(topoInline) > 0 {
		g.logger.Info("reading topo inline")
		err := json.Unmarshal([]byte(topoInline), &topoFile)
		if err != nil {
			return nil, fmt.Errorf("could not parse inline json file: %v", err)
		}
		return &topoFile, nil
	}
	g.logger.Info("reading topo from file path", zap.String("path", g.topoPath))
	parsedFile, err := parseTopoFile(path)

	if err != nil {
		return nil, err
	}
	return parsedFile, nil
}

func (g generatorReceiver) Start(ctx context.Context, host component.Host) error {
	topoFile, err := g.loadTopoFile(g.topoInline, g.topoPath)
	if err != nil {
		host.ReportFatalError(err)
	}

	if g.metricConsumer != nil {
		for _, s := range topoFile.Topology.Services {
			for _, m := range s.Metrics {
				metricTicker := time.NewTicker(1 * time.Second)
				g.tickers = append(g.tickers, metricTicker)
				metricDone := make(chan bool)
				svc := s.ServiceName
				metricName := m.Name
				metricType := m.Type
				go func() {
					g.logger.Info("generating metrics", zap.String("service", svc), zap.String("name", metricName))
					metricGen := generator.NewMetricGenerator(g.randomSeed)
					for {
						select {
						case <-metricDone:
							return
						case _ = <-metricTicker.C:
							metrics := metricGen.Generate(metricName, metricType, svc)
							err := g.metricConsumer.ConsumeMetrics(ctx, metrics)
							if err != nil {
								host.ReportFatalError(err)
							}
						}
					}
				}()
			}
		}

	}
	if g.traceConsumer != nil {
		for _, r := range topoFile.RootRoutes {
			traceTicker := time.NewTicker(time.Duration(360000/r.TracesPerHour) * time.Millisecond)
			g.tickers = append(g.tickers, traceTicker)
			done := make(chan bool)
			svc := r.Service
			route := r.Route
			go func() {
				g.logger.Info("generating traces", zap.String("service", svc), zap.String("route", route))
				traceGen := generator.NewTraceGenerator(topoFile.Topology, g.randomSeed, svc, route)
				for {
					select {
					case <-done:
						return
					case _ = <-traceTicker.C:
						traces := traceGen.Generate(time.Now().UnixNano())
						_ = g.traceConsumer.ConsumeTraces(context.Background(), *traces)
					}
				}
			}()
		}
	}

	return nil
}

var genReceiver = generatorReceiver{}

func (g generatorReceiver) Shutdown(ctx context.Context) error {
	for _, t := range g.tickers {
		t.Stop()
	}
	return nil
}

func newMetricReceiver(config *Config,
	consumer consumer.Metrics,
	logger *zap.Logger, randomSeed int64) (component.MetricsReceiver, error) {

	if consumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}

	genReceiver.logger = logger
	genReceiver.topoPath = config.Path
	genReceiver.topoInline = config.InlineFile
	genReceiver.randomSeed = randomSeed
	genReceiver.metricConsumer = consumer
	return &genReceiver, nil
}

func newTraceReceiver(config *Config,
	consumer consumer.Traces,
	logger *zap.Logger, randomSeed int64) (component.TracesReceiver, error) {

	if consumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}

	genReceiver.logger = logger
	genReceiver.topoPath = config.Path
	genReceiver.randomSeed = randomSeed
	genReceiver.traceConsumer = consumer
	return &genReceiver, nil
}
