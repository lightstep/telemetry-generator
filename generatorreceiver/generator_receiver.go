package generatorreceiver

import (
	"context"
	"encoding/json"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/generator"
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenterror"
	"go.opentelemetry.io/collector/consumer"
	"go.uber.org/zap"
	"io/ioutil"
	"os"
	"time"
)

type generatorReceiver struct {
	logger     *zap.Logger
	consumer   consumer.Traces
	topoPath string
	topoFile topology.File
	traceGen *generator.TraceGenerator
}

func (g generatorReceiver) Start(ctx context.Context, host component.Host) error {
	jsonFile, err := os.Open(g.topoPath)
	if err != nil {
		host.ReportFatalError(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	err = json.Unmarshal(byteValue, &g.topoFile)
	if err != nil {
		host.ReportFatalError(err)
	}
	// TODO: allow generator to be seeded
	g.traceGen = generator.NewTraceGenerator(g.topoFile.Topology, time.Now().Unix())
	for _, r := range g.topoFile.RootRoutes {
		ticker := time.NewTicker(time.Duration(360000/r.TracesPerHour) * time.Millisecond)
		done := make(chan bool)

		go func() {
			for {
				select {
				case <-done:
					return
				case _ = <-ticker.C:
					traces := g.traceGen.Generate(r.Service, r.Route, time.Now().UnixNano())
					_ = g.consumer.ConsumeTraces(context.Background(), *traces)
				}
			}
		}()

	}
	return nil
}

func (g generatorReceiver) Shutdown(ctx context.Context) error {
	// TODO: stop all tickers
	return nil
}

// https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/receiver/awsxrayreceiver/internal/translator/translator.go#L37

func newReceiver(config *Config,
	consumer consumer.Traces,
	logger *zap.Logger) (component.TracesReceiver, error) {

	if consumer == nil {
		return nil, componenterror.ErrNilNextConsumer
	}

	return &generatorReceiver{
		logger: logger,
		topoPath: config.Path,
		consumer: consumer,
	}, nil
}
