package generator

import (
	"github.com/lightstep/telemetry-generator/generatorreceiver/internal/topology"
	"time"

	"go.opentelemetry.io/collector/model/pdata"
)

type MetricGenerator struct {
	metricCount int
}

func NewMetricGenerator() *MetricGenerator {
	return &MetricGenerator{
		metricCount: 0,
	}
}

func (g *MetricGenerator) Generate(metric *topology.Metric, serviceName string) (pdata.Metrics, bool) {
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
		for k, v := range metric.GetTags() {
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
		for k, v := range metric.GetTags() {
			dp.Attributes().UpsertString(k, v)
		}
	}
	// TODO: support histograms!

	g.metricCount = g.metricCount + 1
	return metrics, true
}
