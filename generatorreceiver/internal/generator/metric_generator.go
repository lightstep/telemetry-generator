package generator

import (
	"github.com/lightstep/lightstep-partner-sdk/collector/generatorreceiver/internal/topology"
	"math/rand"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

type MetricGenerator struct {
	metricCount int
	random      *rand.Rand
}

func NewMetricGenerator(seed int64) *MetricGenerator {
	r := rand.New(rand.NewSource(seed))
	r.Seed(seed)
	return &MetricGenerator{
		metricCount: 0,
		random:      r,
	}
}

func (g *MetricGenerator) Generate(metric topology.Metric, serviceName string) (pdata.Metrics, bool) {
	metrics := pdata.NewMetrics()

	if !metric.ShouldGenerate() {
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
		// TODO: support int-type values
		// TODO: support cumulative?
		m.SetDataType(pdata.MetricDataTypeSum)
		m.Sum().SetIsMonotonic(true)
		m.Sum().SetAggregationTemporality(pdata.AggregationTemporalityDelta)
		dp := m.Sum().DataPoints().AppendEmpty()
		dp.SetStartTimestamp(pdata.NewTimestampFromTime(time.Now()))
		dp.SetTimestamp(pdata.NewTimestampFromTime(time.Now()))
		dp.SetDoubleVal(metric.GetValue())
		for k, v := range metric.Tags {
			dp.Attributes().UpsertString(k, v)
		}
	}
	// TODO: support histograms!

	g.metricCount = g.metricCount + 1
	return metrics, true
}
