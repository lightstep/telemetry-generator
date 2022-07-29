package generatorreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/generator"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
)

type generatorReceiver struct {
	logger         *zap.Logger
	traceConsumer  consumer.Traces
	metricConsumer consumer.Metrics
	topoPath       string
	topoInline     string
	randomSeed     int64
	metricGen      *generator.MetricGenerator
	tickers        []*time.Ticker
	fm             *flags.FlagManager
	im             *flags.IncidentManager
	server         *httpServer
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

	g.im = flags.NewIncidentManager()
	g.fm = flags.NewFlagManager(g.im, topoFile.Flags, g.logger)
	g.logger.Info("starting flag manager", zap.Int("flag_count", len(g.fm.Flags)))
	g.fm.Start()

	if g.server != nil {
		err := g.server.Start(ctx, host, g.fm, g.im)
		if err != nil {
			g.logger.Fatal("could not start server", zap.Error(err))
		}
	}

	for _, s := range topoFile.Topology.Services {
		for _, resource := range s.ResourceAttributeSets {
			resource.Kubernetes.CreatePod(s)

			for k, v := range resource.Kubernetes.GetK8sTags() {
				resource.ResourceAttributes[k] = v
			}
		}
	}

	if g.metricConsumer != nil {
		for _, s := range topoFile.Topology.Services {

			var effectiveMetrics []topology.Metric

			// All defined metrics
			effectiveMetrics = append(effectiveMetrics, s.Metrics...)

			// K8s generated metrics
			for _, resource := range s.ResourceAttributeSets {
				// For each resource generate k8s metrics if enabled
				k8sMetrics := resource.Kubernetes.GenerateMetrics(s)
				if k8sMetrics != nil {
					effectiveMetrics = append(effectiveMetrics, k8sMetrics...)
				}
			}

			for _, m := range effectiveMetrics {
				metricTicker := time.NewTicker(1 * time.Second)
				g.tickers = append(g.tickers, metricTicker)
				// TODO: this channel should respect shutdown.
				metricDone := make(chan bool)
				go func(s topology.ServiceTier, m topology.Metric) {
					g.logger.Info("generating metrics", zap.String("service", s.ServiceName), zap.String("name", m.Name))
					metricGen := generator.NewMetricGenerator(g.randomSeed, g.fm)
					for {
						select {
						case <-metricDone:
							return
						case _ = <-metricTicker.C:
							if metrics, report := metricGen.Generate(m, s.ServiceName); report {
								err := g.metricConsumer.ConsumeMetrics(ctx, metrics)
								if err != nil {
									host.ReportFatalError(err)
								}
							}
						}
					}
				}(s, m)
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
				traceGen := generator.NewTraceGenerator(topoFile.Topology, g.randomSeed, svc, route, g.fm)
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
	g.fm.Stop()
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

	// TODO: share server between trace and metric pipelines
	if config.ApiIngress.Endpoint != "" {
		server, err := newHTTPServer(config, logger)
		if err != nil {
			logger.Fatal("could not start http server")
		}
		genReceiver.server = server
	}

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
