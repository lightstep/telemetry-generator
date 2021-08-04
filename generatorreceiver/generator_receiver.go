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
	traceGen   *generator.TraceGenerator
	metricGen  *generator.MetricGenerator
	tickers    []*time.Ticker
}

func (g generatorReceiver) Start(ctx context.Context, host component.Host) error {
	topoFile, err := parseTopoFile(g.topoPath)
	if err != nil {
		host.ReportFatalError(err)
	}

	if g.metricConsumer != nil {
		g.metricGen = generator.NewMetricGenerator(g.randomSeed)
		for _, s := range topoFile.Topology.Services {
			for _, m := range s.Metrics {
				metricTicker := time.NewTicker(1 * time.Second)
				g.tickers = append(g.tickers, metricTicker)
				metricDone := make(chan bool)
				go func() {
					for {
						select {
						case <-metricDone:
							return
						case _ = <-metricTicker.C:
							metrics := g.metricGen.Generate(m.Name, m.Type, s.ServiceName)
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
		g.traceGen = generator.NewTraceGenerator(topoFile.Topology, g.randomSeed)
		for _, r := range topoFile.RootRoutes {
			traceTicker := time.NewTicker(time.Duration(360000/r.TracesPerHour) * time.Millisecond)
			g.tickers = append(g.tickers, traceTicker)
			done := make(chan bool)

			go func() {
				for {
					select {
					case <-done:
						return
					case _ = <-traceTicker.C:
						traces := g.traceGen.Generate(r.Service, r.Route, time.Now().UnixNano())
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
