package generatorreceiver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/cron"
	"math/rand"
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
	server         *httpServer
}

func (g generatorReceiver) loadTopoFile(topoInline string, path string) (topoFile *topology.File, err error) {
	// fetch from env var.
	if len(topoInline) > 0 {
		g.logger.Info("reading topo inline")
		err = json.Unmarshal([]byte(topoInline), topoFile)
		if err != nil {
			return nil, fmt.Errorf("could not parse inline json file: %v", err)
		}
	} else {
		g.logger.Info("reading topo from file path", zap.String("path", g.topoPath))
		topoFile, err = parseTopoFile(path)
		if err != nil {
			return nil, err
		}
	}
	err = topoFile.Topology.Load()
	if err != nil {
		return nil, err
	}
	flags.Manager.LoadFlags(topoFile.Flags, g.logger)

	return topoFile, nil
}

func (g generatorReceiver) validateConfiguration(topoFile *topology.File) error {
	err := flags.Manager.ValidateFlags()
	if err != nil {
		return fmt.Errorf("validation of flag configuration failed: %v", err)
	}

	for _, service := range topoFile.Topology.Services {
		err = service.Validate(*topoFile.Topology) //todo- replace this
		if err != nil {
			return fmt.Errorf("validation of service configuration failed: %v", err)
		}
	}

	err = topoFile.Topology.ValidateServiceGraph() // depends on all services/routes being validated (i.e. exist) first ^
	if err != nil {
		return fmt.Errorf("cyclical service graph detected: %v", err)
	}

	err = topoFile.ValidateRootRoutes()
	if err != nil {
		return fmt.Errorf("validation of rootRoute configuration failed: %v", err)
	}

	return nil
}

func (g generatorReceiver) Start(ctx context.Context, host component.Host) error {
	topoFile, err := g.loadTopoFile(g.topoInline, g.topoPath)
	if err != nil {
		host.ReportFatalError(err)
	}

	err = g.validateConfiguration(topoFile) //todo- should I pass the value instead of the pointer?
	if err != nil {
		host.ReportFatalError(err)
	}

	g.logger.Info("starting flag manager", zap.Int("flag_count", flags.Manager.FlagCount()))
	cron.Start()
	r := rand.New(rand.NewSource(g.randomSeed))
	r.Seed(g.randomSeed)

	if g.server != nil {
		err := g.server.Start(ctx, host)
		if err != nil {
			g.logger.Fatal("could not start server", zap.Error(err))
		}
	}

	for _, s := range topoFile.Topology.Services {
		for i := range s.ResourceAttributeSets {
			s.ResourceAttributeSets[i].Kubernetes.CreatePod(*s)

			if s.ResourceAttributeSets[i].ResourceAttributes == nil {
				s.ResourceAttributeSets[i].ResourceAttributes = make(topology.TagMap)
			}

			for k, v := range s.ResourceAttributeSets[i].Kubernetes.GetK8sTags() {
				s.ResourceAttributeSets[i].ResourceAttributes[k] = v
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
				k8sMetrics := resource.Kubernetes.GenerateMetrics(*s)
				if k8sMetrics != nil {

					for i := range k8sMetrics {
						// keep the same flags as the resources.
						k8sMetrics[i].EmbeddedFlags = resource.EmbeddedFlags
					}

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
					metricGen := generator.NewMetricGenerator(r)
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
				}(*s, m)
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

func (g generatorReceiver) Shutdown(_ context.Context) error {
	for _, t := range g.tickers {
		t.Stop()
	}
	cron.Stop()
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
