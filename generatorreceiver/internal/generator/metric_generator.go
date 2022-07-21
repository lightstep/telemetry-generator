package generator

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
	"math/rand"
	"time"

	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/flags"
	"go.opentelemetry.io/collector/model/pdata"
)

type MetricGenerator struct {
	metricCount int
	random      *rand.Rand
	flagManager *flags.FlagManager
}

func NewMetricGenerator(seed int64, fm *flags.FlagManager) *MetricGenerator {
	r := rand.New(rand.NewSource(seed))
	r.Seed(seed)
	return &MetricGenerator{
		metricCount: 0,
		random:      r,
		flagManager: fm,
	}
}

func (g *MetricGenerator) Generate(metric topology.Metric, serviceName string) (pdata.Metrics, bool) {
	metrics := pdata.NewMetrics()

	if !metric.ShouldGenerate(g.flagManager) {
		return metrics, false
	}

	rms := metrics.ResourceMetrics().AppendEmpty()
	rms.Resource().Attributes().InsertString("service.name", serviceName)

	m := rms.InstrumentationLibraryMetrics().AppendEmpty().Metrics().AppendEmpty()
	m.SetName(metric.Name)
	if metric.Type == "Gauge" {
		m.SetDataType(pdata.MetricDataTypeGauge)
		dp := m.Gauge().DataPoints().AppendEmpty()
		dp.SetTimestamp(pdata.NewTimestampFromTime(time.Now()))
		dp.SetDoubleVal(metric.GetValue())
		for k, v := range metric.Tags {
			dp.Attributes().UpsertString(k, v)
		}
	} else if metric.Type == "Sum" {
		// TODO: support histograms instead :-D
		m.SetDataType(pdata.MetricDataTypeSum)
		dp := m.Sum().DataPoints().AppendEmpty()
		dp.SetTimestamp(pdata.NewTimestampFromTime(time.Now()))
		dp.SetDoubleVal(metric.GetValue())
		for k, v := range metric.Tags {
			dp.Attributes().UpsertString(k, v)
		}
	}

	g.metricCount = g.metricCount + 1
	return metrics, true
}
