package generatorreceiver

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/cron"
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/flags"
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/generator"
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/topology"
)

type generatorReceiver struct {
	logger         *zap.Logger
	traceConsumer  consumer.Traces
	metricConsumer consumer.Metrics
	topoPath       string
	topoInline     string
	randomSeed     int64
	tickers        []*time.Ticker
	server         *httpServer
}

func (g generatorReceiver) loadTopoFile(path string) (topoFile *topology.File, err error) {
	g.logger.Info("reading topo from file path", zap.String("path", g.topoPath))
	topoFile, err = parseTopoFile(path)
	if err != nil {
		return nil, err
	}
	flags.Manager.LoadFlags(topoFile.Flags, g.logger)

	err = topoFile.Topology.Load()
	if err != nil {
		return nil, err
	}

	return topoFile, nil
}

func (g generatorReceiver) Start(ctx context.Context, host component.Host) error {
	topoFile, err := g.loadTopoFile(g.topoPath)
	if err != nil {
		return fmt.Errorf("could not load topo file: %w", err)
	}

	err = validateConfiguration(*topoFile)
	if err != nil {
		return fmt.Errorf("could not validate topo file: %w", err)
	}

	g.logger.Info("starting flag manager", zap.Int("flag_count", flags.Manager.FlagCount()))
	cron.Start()

	// rand is used to generate seeds the underlying *rand.Rand
	generatorRand := rand.New(rand.NewSource(g.randomSeed))

	// Metrics generator uses the global rand.Rand
	rand.Seed(generatorRand.Int63())

	if g.server != nil {
		err := g.server.Start(ctx, host)
		if err != nil {
			g.logger.Fatal("could not start server", zap.Error(err))
		}
	}

	for _, s := range topoFile.Topology.Services {
		for i := range s.ResourceAttributeSets {
			k := s.ResourceAttributeSets[i].Kubernetes
			if k == nil {
				continue
			}
			k.Cfg = topoFile.Config
			k.CreatePods(s.ServiceName)
		}
	}

	if g.metricConsumer != nil {
		for _, s := range topoFile.Topology.Services {

			// Service defined metrics
			for _, m := range s.Metrics {
				metricTicker := g.startMetricGenerator(ctx, s.ServiceName, m)
				g.tickers = append(g.tickers, metricTicker)
			}

			// Service kubernetes auto-generated metrics
			for i := range s.ResourceAttributeSets {
				resource := &s.ResourceAttributeSets[i]
				// For each resource generate k8s metrics if enabled
				if resource.Kubernetes == nil {
					continue
				}
				k8sMetrics := resource.Kubernetes.GenerateMetrics()
				for i := range k8sMetrics {
					// keep the same flags as the resources.
					k8sMetrics[i].EmbeddedFlags = resource.EmbeddedFlags

					metricTicker := g.startMetricGenerator(ctx, s.ServiceName, k8sMetrics[i])
					g.tickers = append(g.tickers, metricTicker)
				}
			}
		}

	}
	if g.traceConsumer != nil {
		for _, rootRoute := range topoFile.RootRoutes {
			traceTicker := time.NewTicker(time.Duration(360000/rootRoute.TracesPerHour) * time.Millisecond)
			g.tickers = append(g.tickers, traceTicker)
			done := make(chan bool)
			svc := rootRoute.Service
			route := rootRoute.Route
			rootRoute := rootRoute

			// rand.Rand is not safe to use in different go routines,
			// create one for each go routine, but use the generatorRand to
			// generate the seed.
			routeRand := rand.New(rand.NewSource(generatorRand.Int63()))

			go func() {
				g.logger.Info("generating traces", zap.String("service", svc), zap.String("route", route))
				traceGen := generator.NewTraceGenerator(topoFile.Topology, routeRand, svc, route)
				for {
					select {
					case <-done:
						return
					case <-traceTicker.C:
						if rootRoute.ShouldGenerate() {
							traces := traceGen.Generate(time.Now().UnixNano())
							err := g.traceConsumer.ConsumeTraces(context.Background(), *traces)
							if err != nil {
								g.logger.Error("consume error", zap.Error(err))
							}

						}
					}
				}
			}()
		}
	}

	return nil
}

func (g *generatorReceiver) startMetricGenerator(ctx context.Context, serviceName string, m topology.Metric) *time.Ticker {
	// TODO: do we actually need to generate every second?
	metricTicker := time.NewTicker(topology.DefaultMetricTickerPeriod)
	go func() {
		g.logger.Info("generating metrics", zap.String("service", serviceName), zap.String("name", m.Name), zap.String("flag_set", m.EmbeddedFlags.FlagSet), zap.String("flag_unset", m.EmbeddedFlags.FlagUnset))
		metricGen := generator.NewMetricGenerator(g.randomSeed)
		for range metricTicker.C {
			m.Pod.RestartIfNeeded(m.EmbeddedFlags, g.logger)

			if metrics, report := metricGen.Generate(&m, serviceName); report {
				err := g.metricConsumer.ConsumeMetrics(ctx, metrics)
				if err != nil {
					g.logger.Error("consume error", zap.Error(err))
				}
			}
		}
	}()

	return metricTicker
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
	logger *zap.Logger, randomSeed int64) (receiver.Metrics, error) {

	if consumer == nil {
		return nil, component.ErrNilNextConsumer
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
	logger *zap.Logger, randomSeed int64) (receiver.Traces, error) {

	if consumer == nil {
		return nil, component.ErrNilNextConsumer
	}

	genReceiver.logger = logger
	genReceiver.topoPath = config.Path
	genReceiver.randomSeed = randomSeed
	genReceiver.traceConsumer = consumer
	return &genReceiver, nil
}

func validateConfiguration(topoFile topology.File) error {
	err := flags.Manager.ValidateFlags()
	if err != nil {
		return fmt.Errorf("validation of flag configuration failed: %v", err)
	}

	for _, service := range topoFile.Topology.Services {
		err = service.Validate(*topoFile.Topology)
		if err != nil {
			return fmt.Errorf("validation of service configuration failed: %v", err)
		}
	}
	err = topoFile.ValidateRootRoutes()
	if err != nil {
		return fmt.Errorf("validation of rootRoute configuration failed: %v", err)
	}

	err = topoFile.Topology.ValidateServiceGraph(topoFile.RootRoutes) // depends on all services/routes being validated (i.e. exist) first
	if err != nil {
		return fmt.Errorf("cyclical service graph detected: %v", err)
	}

	return nil
}
