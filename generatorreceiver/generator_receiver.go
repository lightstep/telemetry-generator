package generatorreceiver

import (
	"context"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/generator"
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
	randomSeed int64
	metricGen  *generator.MetricGenerator
	tickers    []*time.Ticker
}

func (g generatorReceiver) Start(ctx context.Context, host component.Host) error {
	topoFile, err := parseTopoFile(g.topoPath)
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
